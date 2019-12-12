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

package meekserver

import (
	"errors"
	"net"
	"time"
)

// Transport that uses domain fronting to shapeshift the application network traffic
type meekServer struct {
	disableTLS   bool
	certFilename string
	keyFilename  string
}

type meekListener struct {
	listener net.Listener
}

func (m meekListener) Accept() (net.Conn, error) {
	conn, err := m.listener.Accept()
	if err != nil {
		return nil, err
	}

	return meekServerConn{conn}, nil
}

func (m meekListener) Close() error {
	return m.listener.Close()
}

func (m meekListener) Addr() net.Addr {
	interfaces, _ := net.Interfaces()
	addrs, _ := interfaces[0].Addrs()
	return addrs[0]
}
 type meekServerConn struct {
 	conn net.Conn
 }

func (o meekServerConn) Read(b []byte) (n int, err error) {
	return 0, errors.New("unimplemented")
}

func (o meekServerConn) Write(b []byte) (n int, err error) {
	return 0, errors.New("unimplemented")
}

func (o meekServerConn) Close() error {
	return errors.New("unimplemented")
}

func (o meekServerConn) LocalAddr() net.Addr {
	return nil
}

func (o meekServerConn) RemoteAddr() net.Addr {
	return nil
}

func (o meekServerConn) SetDeadline(t time.Time) error {
	return errors.New("unimplemented")
}

func (o meekServerConn) SetReadDeadline(t time.Time) error {
	return errors.New("unimplemented")
}

func (o meekServerConn) SetWriteDeadline(t time.Time) error {
	return errors.New("unimplemented")
}

// Public initializer method to get a new meek transport

func NewMeekServer(disableTLS bool, certFilename string, keyFilename string) *meekServer {
	if disableTLS {
		if certFilename != "" || keyFilename != "" {
			return nil
		}
	} else {
		if certFilename == "" || keyFilename == "" {
			return nil
		}
	}
	return &meekServer{disableTLS, certFilename, keyFilename}
}

// Methods that implement the base.Transport interface

// The meek transport does not have a corresponding server, only a client
func (transport *meekServer) Listen(address string) net.Listener {
	var ln net.Listener
	var err error
	addr, resolverr := net.ResolveTCPAddr("tcp", address)
	if resolverr != nil {
		return ln
	}
	if transport.disableTLS {
		ln, err = startListener("tcp", addr)
	} else {
		ln, err = startListenerTLS("tcp", addr, transport.certFilename, transport.keyFilename)
	}
	if err != nil {
		return ln
	}

	return meekListener{ln}
}

// End methods that implement the base.Transport interface

//func (ca *meekClientArgs) String() string {
//	return "meek" + ":" + ca.front + ":" + ca.url.String()
//}
//
//func newClientArgs(url string) (ca *meekClientArgs, err error) {
//	ca = &meekClientArgs{}
//
//	// Parse the URL argument.
//	ca.url, err = gourl.Parse(url)
//	if err != nil {
//		return nil, fmt.Errorf("malformed url: '%s'", url)
//	}
//	switch ca.url.Scheme {
//	case "http", "https":
//	default:
//		return nil, fmt.Errorf("invalid scheme: '%s'", ca.url.Scheme)
//	}
//
//	return ca, nil
//}
//
//func newClientArgsWithFront(url string, front string) (ca *meekClientArgs, err error) {
//	ca = &meekClientArgs{}
//
//	// Parse the URL argument.
//	ca.url, err = gourl.Parse(url)
//	if err != nil {
//		return nil, fmt.Errorf("malformed url: '%s'", url)
//	}
//	switch ca.url.Scheme {
//	case "http", "https":
//	default:
//		return nil, fmt.Errorf("invalid scheme: '%s'", ca.url.Scheme)
//	}
//
//	// Parse the (optional) front argument.
//	ca.front = front
//
//	return ca, nil
//}
//
//// Implementation of base.TransportConn, which also requires implementing net.Conn
//type meekConn struct {
//	sync.Mutex
//
//	args      *meekClientArgs
//	sessionID string
//	transport *http.Transport
//
//	workerRunning   bool
//	workerWrChan    chan []byte
//	workerRdChan    chan []byte
//	workerCloseChan chan bool
//	rdBuf           *bytes.Buffer
//}
//
//// Private initializer methods
//
//func newMeekClientConn(addr string, ca *meekClientArgs) (*meekConn, error) {
//	id, err := newSessionID()
//	if err != nil {
//		return nil, err
//	}
//
//	tr := &http.Transport{}
//	conn := &meekConn{
//		args:            ca,
//		sessionID:       id,
//		transport:       tr,
//		workerRunning:   true,
//		workerWrChan:    make(chan []byte, maxChanBacklog),
//		workerRdChan:    make(chan []byte, maxChanBacklog),
//		workerCloseChan: make(chan bool),
//	}
//
//	// Start the I/O worker.
//	go conn.ioWorker()
//
//	return conn, nil
//}
//
//// End initializer methods
//// End methods that implement the base.TransportConn interface
//
//// Methods implementing net.Conn
//func (c *meekConn) Read(p []byte) (n int, err error) {
//	// If there is data left over from the previous read,
//	// service the request using the buffered data.
//	if c.rdBuf != nil {
//		if c.rdBuf.Len() == 0 {
//			panic("empty read buffer")
//		}
//		n, err = c.rdBuf.Read(p)
//		if c.rdBuf.Len() == 0 {
//			c.rdBuf = nil
//		}
//		return
//	}
//
//	// Wait for the worker to enqueue more incoming data.
//	b, ok := <-c.workerRdChan
//	if !ok {
//		// Close() was called and the worker's shutting down.
//		return 0, io.ErrClosedPipe
//	}
//
//	// Ew, an extra copy, but who am I kidding, it's meek.
//	buf := bytes.NewBuffer(b)
//	n, err = buf.Read(p)
//	if buf.Len() > 0 {
//		// If there's data pending, stash the buffer so the next
//		// Read() call will use it to fulfuill the Read().
//		c.rdBuf = buf
//	}
//	return
//}
//
//func (c *meekConn) Write(b []byte) (n int, err error) {
//	// Check to see if the connection is actually open.
//	c.Lock()
//	closed := !c.workerRunning
//	c.Unlock()
//	if closed {
//		return 0, io.ErrClosedPipe
//	}
//
//	if len(b) == 0 {
//		return 0, nil
//	}
//
//	// Copy the data to be written to a new slice, since
//	// we return immediately after queuing and the peer can
//	// happily reuse `b` before data has been sent.
//	toWrite := len(b)
//	b2 := make([]byte, toWrite)
//	copy(b2, b)
//	if ok := c.enqueueWrite(b2); !ok {
//		// Technically we did enqueue data, but the worker's
//		// got closed out from under us.
//		return 0, io.ErrClosedPipe
//	}
//	runtime.Gosched()
//	return len(b), nil
//}
//
//func (c *meekConn) Close() error {
//	// Ensure that we do this once and only once.
//	c.Lock()
//	defer c.Unlock()
//	if !c.workerRunning {
//		return nil
//	}
//
//	// Tear down the worker.
//	c.workerRunning = false
//	c.workerCloseChan <- true
//
//	return nil
//}
//
//func (c *meekConn) LocalAddr() net.Addr {
//	return &net.IPAddr{IP: loopbackAddr}
//}
//
//func (c *meekConn) RemoteAddr() net.Addr {
//	return c.args
//}
//
//func (c *meekConn) SetDeadline(t time.Time) error {
//	return ErrNotSupported
//}
//
//func (c *meekConn) SetReadDeadline(t time.Time) error {
//	return ErrNotSupported
//}
//
//func (c *meekConn) SetWriteDeadline(t time.Time) error {
//	return ErrNotSupported
//}
//
//// End of methods implementing net.Conn
//// Force type checks to make sure that instances conform to interfaces
//var _ net.Conn = (*meekConn)(nil)
//var _ net.Addr = (*meekClientArgs)(nil)
