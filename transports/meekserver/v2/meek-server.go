// Package meekserver is the server transport plugin for the meek pluggable transport.
// It acts as an HTTP server, keeps track of session ids, and forwards received
// data to a local OR port.
//
// Sample usage in torrc:
// 	ServerTransportListenAddr meek 0.0.0.0:443
// 	ServerTransportPlugin meek exec ./meek-server --acme-hostnames meek-server.example --acme-email admin@meek-server.example --log meek-server.log
// Using your own TLS certificate:
// 	ServerTransportListenAddr meek 0.0.0.0:8443
// 	ServerTransportPlugin meek exec ./meek-server --cert cert.pem --key key.pem --log meek-server.log
// Plain HTTP usage:
// 	ServerTransportListenAddr meek 0.0.0.0:8080
// 	ServerTransportPlugin meek exec ./meek-server --disable-tls --log meek-server.log
//
// The server runs in HTTPS mode by default, getting certificates from Let's
// Encrypt automatically. The server opens an auxiliary ACME listener on port 80
// in order for the automatic certificates to work. If you have your own
// certificate, use the --cert and --key options. Use --disable-tls option to
// run with plain HTTP.
package meekserver

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/deckarep/golang-set"
	"golang.org/x/net/http2"
)

const (
	// Reject session ids shorter than this, as a weak defense against
	// client bugs that send an empty session id or something similarly
	// likely to collide.
	minSessionIDLength = 8
	// The largest request body we are willing to process, and the largest
	// chunk of data we'll send back in a response.
	maxPayloadLength = 0x10000
	// Passed as ReadTimeout and WriteTimeout when constructing the
	// http.Server.
	readWriteTimeout = 20 * time.Second
	// Cull unused session ids (with their corresponding OR port connection)
	// if we haven't seen any activity for this long.
	maxSessionStaleness = 120 * time.Second
	// How long to wait for ListenAndServe or ListenAndServeTLS to return an
	// error before deciding that it's not going to return.
	listenAndServeErrorTimeout = 100 * time.Millisecond
)

func httpBadRequest(w http.ResponseWriter) {
	http.Error(w, "Bad request.", http.StatusBadRequest)
}

func httpInternalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal server error.", http.StatusInternalServerError)
}

// Session id maps to an existing OR port connection, which we keep open
// between received requests. The first time we see a new session id, we create
// a new OR port connection.
type Session struct {
	Or       fakeConn
	LastSeen time.Time
}

// Touch marks a session as having been seen just now.
func (session *Session) Touch() {
	session.LastSeen = time.Now()
}

// IsExpired finds out if this session is old enough to be culled
func (session *Session) IsExpired() bool {
	return time.Since(session.LastSeen) > maxSessionStaleness
}

// State serves as the http handler
// There is one state per HTTP listener. In the usual case there is just one
// listener, so there is just one global state.
type State struct {
	sessionMap map[string]*Session
	lock       sync.Mutex
	availableSessions mapset.Set
}

// NewState makes a new state
func NewState() *State {
	state := new(State)
	state.sessionMap = make(map[string]*Session)
	state.availableSessions = mapset.NewSet()
	return state
}

func (state *State) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		state.Get(w, req)
	case "POST":
		state.Post(w, req)
	default:
		httpBadRequest(w)
	}
}

// Get handles a GET request. This doesn't have any purpose apart from diagnostics.
func (state *State) Get(w http.ResponseWriter, req *http.Request) {
	if path.Clean(req.URL.Path) != "/" {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("I’m just a happy little web server.\n"))
}

// Get a string representing the original client address, if available, as a
// "host:port" string suitable to pass as the addr parameter to pt.DialOr. Never
// fails: if the original client address is not available, returns "". If the
// original client address is available, the returned port number is always 1.

// GetSession looks up a session by id, or create a new one (with its OR port connection) if
// it doesn't already exist.
func (state *State) GetSession(sessionID string) (*Session, error) {
	state.lock.Lock()
	defer state.lock.Unlock()

	session := state.sessionMap[sessionID]
	if session == nil {
		// log.Printf("unknown session id %q; creating new session", sessionID)

		session = &Session{Or: newFakeConn()}
		state.sessionMap[sessionID] = session
		state.availableSessions.Add(sessionID)
	}
	session.Touch()

	return session, nil
}

func newFakeConn() fakeConn {
	readBuffer := make([]byte, 0)
	writeBuffer := make([]byte, 0)
	return fakeConn{
		readBuffer:  readBuffer,
		writeBuffer: writeBuffer,
	}
}

// scrubbedAddr is a phony net.Addr that returns "[scrubbed]" for all calls.
type scrubbedAddr struct{}

func (a scrubbedAddr) Network() string {
	return "[scrubbed]"
}
func (a scrubbedAddr) String() string {
	return "[scrubbed]"
}

// Replace the Addr in a net.OpError with "[scrubbed]" for logging.
func scrubError(err error) error {
	if operr, ok := err.(*net.OpError); ok {
		// net.OpError contains Op, Net, Addr, and a subsidiary Err. The
		// (Op, Net, Addr) part is responsible for error text prefixes
		// like "read tcp X.X.X.X:YYYY:". We want that information but
		// don't want to log the literal address.
		operr.Addr = scrubbedAddr{}
	}
	return err
}

// Feed the body of req into the OR port, and write any data read from the OR
// port back to w.
func transact(session *Session, w http.ResponseWriter, req *http.Request) error {
	body := http.MaxBytesReader(w, req.Body, maxPayloadLength+1)
	_, err := io.ReadFull(body, session.Or.readBuffer)
	if err != nil {
		return fmt.Errorf("error copying body to ORPort: %s", scrubError(err))
	}

	buf := make([]byte, maxPayloadLength)
	//session.Or.SetReadDeadline(time.Now().Add(turnaroundTimeout))
	n := len(session.Or.writeBuffer)
	if n == 0 {
			httpInternalServerError(w)
			// Don't scrub err here because it always refers to localhost.
			return fmt.Errorf("reading from ORPort: %s", err)
	}
	if n > maxPayloadLength {
		copy(buf, session.Or.writeBuffer[:maxPayloadLength])
		session.Or.writeBuffer = session.Or.writeBuffer[maxPayloadLength:]
	} else {
		copy(buf, session.Or.writeBuffer)
		session.Or.writeBuffer = session.Or.writeBuffer[:0]
	}
	// log.Printf("read %d bytes from ORPort: %q", n, buf[:n])
	// Set a Content-Type to prevent Go and the CDN from trying to guess.
	w.Header().Set("Content-Type", "application/octet-stream")
	n, err = w.Write(buf)
	if err != nil {
		return fmt.Errorf("error writing to response: %s", scrubError(err))
	}
	// log.Printf("wrote %d bytes to response", n)
	return nil
}

// Post handles a POST request. Look up the session id and then do a transaction.
func (state *State) Post(w http.ResponseWriter, req *http.Request) {
	sessionID := req.Header.Get("X-Session-Id")
	if len(sessionID) < minSessionIDLength {
		httpBadRequest(w)
		return
	}

	session, err := state.GetSession(sessionID)
	if err != nil {
		log.Print(err)
		httpInternalServerError(w)
		return
	}

	err = transact(session, w, req)
	if err != nil {
		log.Print(err)
		state.CloseSession(sessionID)
		return
	}
}

// CloseSession removes a session from the map and closes its corresponding OR port
// connection. Does nothing if the session id is not known.
func (state *State) CloseSession(sessionID string) {
	state.lock.Lock()
	defer state.lock.Unlock()
	// log.Printf("closing session %q", sessionID)
	_, ok := state.sessionMap[sessionID]
	if ok {
		delete(state.sessionMap, sessionID)
		state.availableSessions.Remove(sessionID)
	}
}

// ExpireSessions prevents an endless loop, checking for expired sessions and removing them.
func (state *State) ExpireSessions() {
	for {
		time.Sleep(maxSessionStaleness / 2)
		state.lock.Lock()
		for sessionID, session := range state.sessionMap {
			if session.IsExpired() {
				// log.Printf("deleting expired session %q", sessionID)
				delete(state.sessionMap, sessionID)
				state.availableSessions.Remove(sessionID)
			}
		}
		state.lock.Unlock()
	}
}

func initServer(addr *net.TCPAddr,
	getCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error),
	listenAndServe func(*http.Server, chan<- error)) (*http.Server, *State, error) {
	// We're not capable of listening on port 0 (i.e., an ephemeral port
	// unknown in advance). The reason is that while the net/http package
	// exposes ListenAndServe and ListenAndServeTLS, those functions never
	// return, so there's no opportunity to find out what the port number
	// is, in between the Listen and Serve steps.
	// https://groups.google.com/d/msg/Golang-nuts/3F1VRCCENp8/3hcayZiwYM8J
	if addr.Port == 0 {
		return nil, nil, fmt.Errorf("cannot listen on port %d; configure a port using ServerTransportListenAddr", addr.Port)
	}

	state := NewState()
	go state.ExpireSessions()

	server := &http.Server{
		Addr:         addr.String(),
		Handler:      state,
		ReadTimeout:  readWriteTimeout,
		WriteTimeout: readWriteTimeout,
	}
	// We need to override server.TLSConfig.GetCertificate--but first
	// server.TLSConfig needs to be non-nil. If we just create our own new
	// &tls.Config, it will lack the default settings that the net/http
	// package sets up for things like HTTP/2. Therefore we first call
	// http2.ConfigureServer for its side effect of initializing
	// server.TLSConfig properly. An alternative would be to make a dummy
	// net.Listener, call Serve on it, and let it return.
	// https://github.com/golang/go/issues/16588#issuecomment-237386446
	err := http2.ConfigureServer(server, nil)
	if err != nil {
		return server, state, err
	}
	server.TLSConfig.GetCertificate = getCertificate

	// Another unfortunate effect of the inseparable net/http ListenAndServe
	// is that we can't check for Listen errors like "permission denied" and
	// "address already in use" without potentially entering the infinite
	// loop of Serve. The hack we apply here is to wait a short time,
	// listenAndServeErrorTimeout, to see if an error is returned (because
	// it's better if the error message goes to the tor log through
	// SMETHOD-ERROR than if it only goes to the meek-server log).
	errChan := make(chan error)
	go listenAndServe(server, errChan)
	select {
	case err = <-errChan:
		break
	case <-time.After(listenAndServeErrorTimeout):
		break
	}

	return server, state, err
}

func startServer(addr *net.TCPAddr) (*http.Server, *State, error) {
	return initServer(addr, nil, func(server *http.Server, errChan chan<- error) {
		log.Printf("listening with plain HTTP on %s", addr)
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("Error in ListenAndServe: %s", err)
		}
		errChan <- err
	})
}

func startServerTLS(addr *net.TCPAddr, getCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error)) (*http.Server, *State, error) {
	return initServer(addr, getCertificate, func(server *http.Server, errChan chan<- error) {
		log.Printf("listening with HTTPS on %s", addr)
		err := server.ListenAndServeTLS("", "")
		if err != nil {
			log.Printf("Error in ListenAndServeTLS: %s", err)
		}
		errChan <- err
	})
}

func getCertificateCacheDir(stateDir string) (string, error) {

	return filepath.Join(stateDir, "meek-certificate-cache"), nil
}
