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

// Package meeklite provides an implementation of the Meek circumvention
// protocol.  Only a client implementation is provided, and no effort is
// made to normalize the TLS fingerprint.
//
// It borrows quite liberally from the real meek-client code.
package meeklite

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	gourl "net/url"
	"runtime"
	"sync"
	"time"
)

const (
	maxChanBacklog = 16

	// Constants shamelessly stolen from meek-client.go...
	maxPayloadLength       = 0x10000
	initPollInterval       = 100 * time.Millisecond
	maxPollInterval        = 5 * time.Second
	pollIntervalMultiplier = 1.5
	maxRetries             = 10
	retryDelay             = 30 * time.Second
)

var (
	// ErrNotSupported is the error returned for a unsupported operation.
	ErrNotSupported = errors.New("meek_lite: operation not supported")

	loopbackAddr = net.IPv4(127, 0, 0, 1)
)

// MeekTransport that uses domain fronting to shapeshift the application network traffic
type MeekTransport struct {
	dialer proxy.Dialer

	clientArgs *meekClientArgs
}

// NewMeekTransportWithFront is a public initializer method to get a new meek transport
func NewMeekTransportWithFront(url string, front string, dialer proxy.Dialer) *MeekTransport {
	clientArgs, err := newClientArgsWithFront(url, front)
	if err != nil {
		return nil
	}

	return &MeekTransport{dialer: dialer, clientArgs: clientArgs}
}

// Methods that implement the base.Transport interface

// NetworkDialer is a dialer for the underlying network connection
// The Dialer can be modified to change how the network connections are made.
func (transport *MeekTransport) NetworkDialer() proxy.Dialer {
	return transport.dialer
}

// Dial creates outgoing transport connection
func (transport *MeekTransport) Dial() net.Conn {
	// FIXME - should use dialer
	transportConn, err := newMeekClientConn(transport.clientArgs)
	if err != nil {
		return nil
	}

	return transportConn
}

// Listen for the meek transport does not have a corresponding server, only a client
func (transport *MeekTransport) Listen() net.Listener {
	return nil
}

// End methods that implement the base.Transport interface

type meekClientArgs struct {
	url   *gourl.URL
	front string
}

func (ca *meekClientArgs) Network() string {
	return "meek"
}

//Transport contains parameters used in Optimizer
type Transport struct {
	URL     *gourl.URL `json:"url"`
	Front   string     `json:"front"`
	Address string
	Dialer  proxy.Dialer
}

//Config puts the parameters in a json compatible format
type Config struct {
	URL   *gourl.URL `json:"url"`
	Front string     `json:"front"`
}

// Dial creates outgoing transport connection
func (transport Transport) Dial() (net.Conn, error) {
	meekTransport := NewMeekTransportWithFront(transport.URL.String(), transport.Front, transport.Dialer)
	conn := meekTransport.Dial()
	return conn, nil
}

func (ca *meekClientArgs) String() string {
	return "meek" + ":" + ca.front + ":" + ca.url.String()
}

func newClientArgsWithFront(url string, front string) (ca *meekClientArgs, err error) {
	ca = &meekClientArgs{}

	// Parse the URL argument.
	ca.url, err = gourl.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("malformed url: '%s'", url)
	}
	switch ca.url.Scheme {
	case "http", "https":
	default:
		return nil, fmt.Errorf("invalid scheme: '%s'", ca.url.Scheme)
	}

	// Parse the (optional) front argument.
	ca.front = front

	return ca, nil
}

// Implementation of base.TransportConn, which also requires implementing net.Conn
type meekConn struct {
	sync.Mutex

	args      *meekClientArgs
	sessionID string
	transport *http.Transport

	workerRunning   bool
	workerWrChan    chan []byte
	workerRdChan    chan []byte
	workerCloseChan chan bool
	rdBuf           *bytes.Buffer
}

// Private initializer methods

func newMeekClientConn(ca *meekClientArgs) (*meekConn, error) {
	id, err := newSessionID()
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{}
	conn := &meekConn{
		args:            ca,
		sessionID:       id,
		transport:       tr,
		workerRunning:   true,
		workerWrChan:    make(chan []byte, maxChanBacklog),
		workerRdChan:    make(chan []byte, maxChanBacklog),
		workerCloseChan: make(chan bool),
	}

	// Start the I/O worker.
	go conn.ioWorker()

	return conn, nil
}

// End initializer methods

// Methods that implement the base.TransportConn interface
func (transportConn *meekConn) NetworkConn() net.Conn {
	return nil
}

// End methods that implement the base.TransportConn interface

// Methods implementing net.Conn
func (transportConn *meekConn) Read(p []byte) (n int, err error) {
	// If there is data left over from the previous read,
	// service the request using the buffered data.
	if transportConn.rdBuf != nil {
		if transportConn.rdBuf.Len() == 0 {
			panic("empty read buffer")
		}
		n, err = transportConn.rdBuf.Read(p)
		if transportConn.rdBuf.Len() == 0 {
			transportConn.rdBuf = nil
		}
		return
	}
	select {
		case <-time.After(20*time.Second):
			return 0, nil
		// Wait for the worker to enqueue more incoming data.
		case b, ok := <-transportConn.workerRdChan:
			if !ok {
				// Close() was called and the worker's shutting down.
				return 0, io.ErrClosedPipe
			}

			// Ew, an extra copy, but who am I kidding, it's meek.
			buf := bytes.NewBuffer(b)
			n, err = buf.Read(p)
			if buf.Len() > 0 {
				// If there's data pending, stash the buffer so the next
				// Read() call will use it to fulfill the Read().
				transportConn.rdBuf = buf
			}
	}
	return
}

func (transportConn *meekConn) Write(b []byte) (n int, err error) {
	// Check to see if the connection is actually open.
	transportConn.Lock()
	closed := !transportConn.workerRunning
	transportConn.Unlock()
	if closed {
		return 0, io.ErrClosedPipe
	}

	if len(b) == 0 {
		return 0, nil
	}

	// Copy the data to be written to a new slice, since
	// we return immediately after queuing and the peer can
	// happily reuse `b` before data has been sent.
	toWrite := len(b)
	b2 := make([]byte, toWrite)
	copy(b2, b)
	if ok := transportConn.enqueueWrite(b2); !ok {
		// Technically we did enqueue data, but the worker's
		// got closed out from under us.
		return 0, io.ErrClosedPipe
	}
	runtime.Gosched()
	return len(b), nil
}

func (transportConn *meekConn) Close() error {
	// Ensure that we do this once and only once.
	transportConn.Lock()
	defer transportConn.Unlock()
	if !transportConn.workerRunning {
		return nil
	}

	// Tear down the worker.
	transportConn.workerRunning = false
	transportConn.workerCloseChan <- true

	return nil
}

func (transportConn *meekConn) LocalAddr() net.Addr {
	return &net.IPAddr{IP: loopbackAddr}
}

func (transportConn *meekConn) RemoteAddr() net.Addr {
	return transportConn.args
}

func (transportConn *meekConn) SetDeadline(time.Time) error {
	return ErrNotSupported
}

func (transportConn *meekConn) SetReadDeadline(time.Time) error {
	return ErrNotSupported
}

func (transportConn *meekConn) SetWriteDeadline(time.Time) error {
	return ErrNotSupported
}

// End of methods implementing net.Conn

// Private methods implementing the domain fronting cipher
func (transportConn *meekConn) enqueueWrite(b []byte) (ok bool) {
	defer func() { recover() }()
	transportConn.workerWrChan <- b
	return true
}

func (transportConn *meekConn) roundTrip(sndBuf []byte) (recvBuf []byte, err error) {
	var req *http.Request
	var resp *http.Response

	for retries := 0; retries < maxRetries; retries++ {
		url := *transportConn.args.url
		host := url.Host
		if transportConn.args.front != "" {
			url.Host = transportConn.args.front
		}

		req, err = http.NewRequest("POST", url.String(), bytes.NewReader(sndBuf))
		if err != nil {
			return nil, err
		}

		if transportConn.args.front != "" {
			req.Host = host
		}

		req.Header.Set("X-Session-Id", transportConn.sessionID)
		req.Header.Set("User-Agent", "")

		resp, err = transportConn.transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("status code was %d, not %d", resp.StatusCode, http.StatusOK)
			if resp.StatusCode == http.StatusInternalServerError {
				return
			}
				time.Sleep(retryDelay)

		} else {
			_ = resp.Body.Close()
			recvBuf, err = ioutil.ReadAll(io.LimitReader(resp.Body, maxPayloadLength))
			return
		}
	}

	return
}

func (transportConn *meekConn) ioWorker() {
	interval := initPollInterval
	var sndBuf, leftBuf []byte
loop:

	for {
		sndBuf = nil
		select {
		case <-time.After(interval):
			// If the poll interval has elapsed, issue a request.
		case sndBuf = <-transportConn.workerWrChan:
			// If there is data pending a send, issue a request.
		case _ = <-transportConn.workerCloseChan:
			break loop
		}

		// Combine short writes as long as data is available to be
		// sent immediately and it will not put us over the max
		// payload limit.  Any excess data is stored and dispatched
		// as the next request).
		sndBuf = append(leftBuf, sndBuf...)
		wrSz := len(sndBuf)
		for len(transportConn.workerWrChan) > 0 && wrSz < maxPayloadLength {

			b := <-transportConn.workerWrChan
			sndBuf = append(sndBuf, b...)
			wrSz = len(sndBuf)
		}
		if wrSz > maxPayloadLength {
			wrSz = maxPayloadLength
		}

		// Issue a request.
		rdBuf, err := transportConn.roundTrip(sndBuf[:wrSz])
		if err != nil {
			// Welp, something went horrifically wrong.
			break loop
		}

		// Stash the remaining payload if any.
		leftBuf = sndBuf[wrSz:] // Store the remaining data
		if len(leftBuf) == 0 {
			leftBuf = nil
		}

		// Determine the next poll interval.
		if len(rdBuf) > 0 {
			// Received data, enqueue the read.
			transportConn.workerRdChan <- rdBuf

			// And poll immediately.
			interval = 0
		} else if wrSz > 0 {
			// Sent data, poll immediately.
			interval = 0
		} else if interval == 0 {
			// Neither sent nor received data, initialize the delay.
			interval = initPollInterval
		} else {
			// Apply a multiplicative backoff.
			interval = time.Duration(float64(interval) * pollIntervalMultiplier)
			if interval > maxPollInterval {
				interval = maxPollInterval
			}
		}

		runtime.Gosched()
	}

	// Unblock callers waiting in Read() for data that will never arrive,
	// and callers waiting in Write() for data that will never get sent.
	close(transportConn.workerRdChan)
	close(transportConn.workerWrChan)

	// In case the close was done on an error condition, update the state
	// variable so that further calls to Write() will fail.
	transportConn.Lock()
	defer transportConn.Unlock()
	transportConn.workerRunning = false
}

func newSessionID() (string, error) {
	var b [64]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	h := sha256.Sum256(b[:])
	return hex.EncodeToString(h[:16]), nil
}

// End private methods implementing the ROT13 cipher

// Force type checks to make sure that instances conform to interfaces
var _ net.Conn = (*meekConn)(nil)
var _ net.Addr = (*meekClientArgs)(nil)
