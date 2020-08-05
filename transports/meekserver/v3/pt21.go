/*
 * Copyright (c) 2015, Yawning Angel <yawning at torproject dot org>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *  * Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 *
 *  * Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */
/*
 * Copyright (c) 2014, Yawning Angel <yawning at torproject dot org>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *  * Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 *
 *  * Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */


// Package meekserver provides an implementation of the Meek circumvention
// protocol.  Only a client implementation is provided, and no effort is
// made to normalize the TLS fingerprint.
//
// It borrows quite liberally from the real meek-client code.
package meekserver

import (
	"errors"
	"github.com/kataras/golog"
	interconv "github.com/mufti1/interconv/package"
	"golang.org/x/crypto/acme/autocert"
	"net"
	"net/http"
	"strings"
	"time"
)

//MeekServer is a Transport that uses domain fronting to shapeshift the application network traffic
type MeekServer struct {
	DisableTLS   bool
	AcmeEmail    string
	AcmeHostname string
	CertManager  *autocert.Manager
}

type Transport struct {
	DisableTLS   bool
	CertManager  *autocert.Manager
	Address      string
}

//Config contains arguments formatted for a json file
type Config struct {
	AcmeEmail    string `json:"acme-email"`
	AcmeHostname string `json:"acme-hostnames"`
}

type meekListener struct {
	server *http.Server
	state  *State
}

type meekServerConn struct {
	session   *Session
	state     *State
	sessionID string
}

type fakeConn struct {
	readBuffer  []byte
	writeBuffer []byte
}

func New(disableTLS bool, acmeHostnamesCommas string, acmeEmail string, address string, stateDir string) (*Transport, error) {
	var certManager *autocert.Manager
	if disableTLS {
		if acmeEmail != "" || acmeHostnamesCommas != "" {
			return nil, errors.New("acmeEmail and acmeHostnames must be empty when disableTLS is enabled")
		}
		return nil, errors.New("disableTLS mode is not yet supported")
	} else {
		if acmeEmail == "" || acmeHostnamesCommas == "" {
			return nil, errors.New("acmeEmail and acmeHostnames must be empty when disableTLS is disabled")
		}
		if acmeHostnamesCommas != "" {
			acmeHostnames := strings.Split(acmeHostnamesCommas, ",")
			golog.Infof("ACME hostnames: %q", acmeHostnames)

			// The ACME HTTP-01 responder only works when it is running on
			// port 80.
			// https://github.com/ietf-wg-acme/acme/blob/master/draft-ietf-acme-acme.md#http-challenge

			var cache autocert.Cache
			cacheDir, err := getCertificateCacheDir(stateDir)
			if err == nil {
				golog.Infof("caching ACME certificates in directory %q", cacheDir)
				cache = autocert.DirCache(cacheDir)
			} else {
				golog.Infof("disabling ACME certificate cache: %s", err)
			}

			certManager = &autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(acmeHostnames...),
				Email:      acmeEmail,
				Cache:      cache,
			}
			return &Transport{
				DisableTLS:  disableTLS,
				CertManager: certManager,
				Address:     address,
			}, nil
		} else {
			return nil, errors.New("must set acmeEmail")
		}
	}
}
func (listener meekListener) Accept() (net.Conn, error) {
	state := listener.state
	state.lock.Lock()
	defer state.lock.Unlock()
	if state.availableSessions.Cardinality() == 0 {
		return nil, errors.New("no connections available in accept")
	}
	sessionID := state.availableSessions.Pop()
	sessionIDString, err := interconv.ParseString(sessionID)
	if err != nil {
		return nil, errors.New("could not convert sessionID to string")
	}
	return NewMeekServerConnection(state, sessionIDString), nil

}

//NewMeekServerConnection initializes the server connection
func NewMeekServerConnection(state *State, sessionID string) net.Conn {
	session := state.sessionMap[sessionID]
	return meekServerConn{session, state, sessionID}
}

func (listener meekListener) Close() error {
	return listener.server.Close()
}

func (listener meekListener) Addr() net.Addr {
	interfaces, _ := net.Interfaces()
	addrs, _ := interfaces[0].Addrs()
	return addrs[0]
}

func (conn meekServerConn) Read(b []byte) (n int, err error) {
	if len(conn.session.Or.readBuffer) == 0 {
		return 0, nil
	}
	copy(b, conn.session.Or.readBuffer)
	conn.session.Or.readBuffer = conn.session.Or.readBuffer[:0]

	return len(b), nil
}

func (conn meekServerConn) Write(b []byte) (n int, err error) {
	conn.session.Or.writeBuffer = append(conn.session.Or.writeBuffer, b...)
	return len(b), nil
}

func (conn meekServerConn) Close() error {
	conn.state.CloseSession(conn.sessionID)
	return nil
}

//end critical importance
func (conn meekServerConn) LocalAddr() net.Addr {
	return nil
}

func (conn meekServerConn) RemoteAddr() net.Addr {
	return nil
}

func (conn meekServerConn) SetDeadline(time.Time) error {
	return errors.New("unimplemented")
}

func (conn meekServerConn) SetReadDeadline(time.Time) error {
	return errors.New("unimplemented")
}

func (conn meekServerConn) SetWriteDeadline(time.Time) error {
	return errors.New("unimplemented")
}

// NewMeekTransportServer is a public initializer method to get a new meek transport
func NewMeekTransportServer(disableTLS bool, acmeEmail string, acmeHostnamesCommas string, stateDir string) *MeekServer {
	var certManager *autocert.Manager
	if disableTLS {
		if acmeEmail != "" || acmeHostnamesCommas != "" {
			return nil
		}
	} else {
		if acmeEmail == "" || acmeHostnamesCommas == "" {
			return nil
		}
		if acmeHostnamesCommas != "" {
			acmeHostnames := strings.Split(acmeHostnamesCommas, ",")
			golog.Infof("ACME hostnames: %q", acmeHostnames)

			// The ACME HTTP-01 responder only works when it is running on
			// port 80.
			// https://github.com/ietf-wg-acme/acme/blob/master/draft-ietf-acme-acme.md#http-challenge

			var cache autocert.Cache
			cacheDir, err := getCertificateCacheDir(stateDir)
			if err == nil {
				golog.Infof("caching ACME certificates in directory %q", cacheDir)
				cache = autocert.DirCache(cacheDir)
			} else {
				golog.Infof("disabling ACME certificate cache: %s", err)
			}

			certManager = &autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(acmeHostnames...),
				Email:      acmeEmail,
				Cache:      cache,
			}
		}
	}
	return &MeekServer{disableTLS, acmeEmail, acmeHostnamesCommas, certManager}
}

// Methods that implement the base.Transport interface

// Listen on the meek transport does not have a corresponding server, only a client
func (transport *MeekServer) Listen(address string) (net.Listener, error) {
	var ln net.Listener
	var state *State
	var err error
	addr, resolverr := net.ResolveTCPAddr("tcp", address)
	if resolverr != nil {
		return ln, resolverr
	}
	acmeAddr := net.TCPAddr{
		IP:   addr.IP,
		Port: 80,
		Zone: "",
	}
	acmeAddr.Port = 80
	golog.Infof("starting HTTP-01 ACME listener on %s", acmeAddr.String())
	lnHTTP01, err := net.ListenTCP("tcp", &acmeAddr)
	if err != nil {
		golog.Infof("error opening HTTP-01 ACME listener: %s", err)
		return nil, err
	}
	go func() {
		golog.Fatal(http.Serve(lnHTTP01, transport.CertManager.HTTPHandler(nil)))
	}()
	var server *http.Server
	if transport.DisableTLS {
		server, state, err = startServer(addr)
	} else {
		server, state, err = startServerTLS(addr, transport.CertManager.GetCertificate)
	}
	if err != nil {

		return nil, err
	}
	return meekListener{server, state}, nil
}

func (transport *Transport) Listen() (net.Listener, error) {
	var ln net.Listener
	var state *State
	var err error
	addr, resolverr := net.ResolveTCPAddr("tcp", transport.Address)
	if resolverr != nil {
		return ln, resolverr
	}
	acmeAddr := net.TCPAddr{
		IP:   addr.IP,
		Port: 80,
		Zone: "",
	}
	acmeAddr.Port = 80
	golog.Infof("starting HTTP-01 ACME listener on %s", acmeAddr.String())
	lnHTTP01, err := net.ListenTCP("tcp", &acmeAddr)
	if err != nil {
		golog.Infof("error opening HTTP-01 ACME listener: %s", err)
		return nil, err
	}
	go func() {
		golog.Fatal(http.Serve(lnHTTP01, transport.CertManager.HTTPHandler(nil)))
	}()
	var server *http.Server
	if transport.DisableTLS {
		server, state, err = startServer(addr)
	} else {
		server, state, err = startServerTLS(addr, transport.CertManager.GetCertificate)
	}
	if err != nil {

		return nil, err
	}
	return meekListener{server, state}, nil
}
