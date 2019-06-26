/*
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

// Package rot13 provides an an example transport that used the trivial ROT13
// cipher. https://en.wikipedia.org/wiki/ROT13
// This transport should never be used for real traffic.
package rot13

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/proxy"

	"github.com/OperatorFoundation/shapeshifter-ipc"
)

// Transport that uses a ROT13 cipher to shapeshift the application network traffic
type rot13Transport struct {
	dialer *net.Dialer
}

// Public initializer method to get a new ROT13 transport
func NewRot13Transport() *rot13Transport {
	return &rot13Transport{dialer: nil}
}

// Methods that implement the base.Transport interface

// Dialer for the underlying network connection
// The Dialer can be modified to change how the network connections are made.
func (transport *rot13Transport) NetworkDialer() net.Dialer {
	return *transport.dialer
}

// Create outgoing transport connection
func (transport *rot13Transport) Dial(address string) net.Conn {
	// FIXME - should use dialer
	dialFn := proxy.Direct.Dial
	conn, dialErr := dialFn("tcp", address)
	if dialErr != nil {
		return nil
	}

	dialConn := conn
	transportConn, err := newRot13ClientConn(conn)
	if err != nil {
		dialConn.Close()
		return nil
	}

	return transportConn
}

// Create listener for incoming transport connection
func (transport *rot13Transport) Listen(address string) net.Listener {
	addr, resolveErr := pt.ResolveAddr(address)
	if resolveErr != nil {
		fmt.Println(resolveErr.Error())
		return nil
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return newRot13TransportListener(ln)
}

// End methods that implement the base.Transport interface

// Listener that accepts connections using the ROT13 transport to communicate
type rot13TransportListener struct {
	listener *net.TCPListener
}

// Private initializer for the ROT13 listener.
// You get a new listener instance by calling the Listen method on the Transport.
func newRot13TransportListener(listener *net.TCPListener) *rot13TransportListener {
	return &rot13TransportListener{listener: listener}
}

// Methods that implement the net.Listener interface

// Listener for underlying network connection
func (listener *rot13TransportListener) NetworkListener() net.Listener {
	return listener.listener
}

// Accept waits for and returns the next connection to the listener.
func (listener *rot13TransportListener) Accept() (net.Conn, error) {
	conn, err := listener.listener.Accept()
	if err != nil {
		return nil, err
	}

	return newRot13ServerConn(conn)
}

func (listener *rot13TransportListener) Addr() net.Addr {
	interfaces, _ := net.Interfaces()
	addrs, _ := interfaces[0].Addrs()
	return addrs[0]
}

// Close closes the transport listener.
// Any blocked Accept operations will be unblocked and return errors.
func (listener *rot13TransportListener) Close() error {
	return listener.listener.Close()
}

// End methods that implement the net.Listener interface

// Implementation of net.Conn, which also requires implementing net.Conn
type rot13Conn struct {
	conn net.Conn
}

// Private initializer methods
func newRot13ClientConn(conn net.Conn) (c *rot13Conn, err error) {
	// Initialize a client connection, and start the handshake timeout.
	return &rot13Conn{conn: conn}, nil
}

func newRot13ServerConn(conn net.Conn) (c *rot13Conn, err error) {
	// Initialize a server connection, and start the handshake timeout.
	return &rot13Conn{conn}, nil
}

// End initializer methods

// Methods that implement the net.Conn interface
func (transportConn *rot13Conn) NetworkConn() net.Conn {
	return transportConn.conn
}

// End methods that implement the net.Conn interface

// Methods implementing net.Conn
func (conn *rot13Conn) Read(b []byte) (int, error) {
	i, err := conn.conn.Read(b)
	if err != nil {
		return 0, err
	}

	unshift(b)
	return i, err
}

func (conn *rot13Conn) Write(b []byte) (int, error) {
	shift(b)
	i, err := conn.conn.Write(b)
	if err != nil {
		return 0, err
	}

	return i, err
}

func (conn *rot13Conn) Close() error {
	return conn.conn.Close()
}

func (conn *rot13Conn) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}

func (conn *rot13Conn) RemoteAddr() net.Addr {
	return conn.conn.RemoteAddr()
}

func (conn *rot13Conn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

func (conn *rot13Conn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

func (conn *rot13Conn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}

// End of methods implementing net.Conn

// Private methods implementing the ROT13 cipher
func shift(bs []byte) {
	for i, _ := range bs {
		var n int
		n = int(bs[i])
		n = (n + 13) % 255
		bs[i] =  byte(n)
	}
}

func unshift(bs []byte) {
	for i, _ := range bs {
		var n int
		n = int(bs[i])
		n = (n + 255 - 13) % 255
		bs[i] =  byte(n)
	}
}

// End private methods implementing the ROT13 cipher

// Force type checks to make sure that instances conform to interfaces
var _ net.Listener = (*rot13TransportListener)(nil)
var _ net.Conn = (*rot13Conn)(nil)
