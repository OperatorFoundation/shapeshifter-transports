/*
 * Copyright (c) 2017, Operator Foundation
 *
 */

// Package shadow provides a PT 2.0 Go API wrapper around the connections used by Shadowsocks
package shadow

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/shadowsocks/shadowsocks-go/shadowsocks"

	"github.com/OperatorFoundation/shapeshifter-ipc"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/base"
)

// shadowTransport is the shadow implementation of the base.Transport interface.
type shadowTransport struct {
	dialer *net.Dialer

	password   string
	cipherName string
}

func NewShadowTransport(password string, cipherName string) *shadowTransport {
	return &shadowTransport{dialer: nil, password: password, cipherName: cipherName}
}

type shadowTransportListener struct {
	listener  *net.TCPListener
	transport *shadowTransport
}

func newShadowTransportListener(listener *net.TCPListener, transport *shadowTransport) *shadowTransportListener {
	return &shadowTransportListener{listener: listener, transport: transport}
}

// Methods that the implement base.Transport interface
// Dialer for the underlying network connection
// The Dialer can be modified to change how the network connections are made.
func (transport *shadowTransport) NetworkDialer() net.Dialer {
	return *transport.dialer
}

// Create outgoing transport connection
func (transport *shadowTransport) Dial(address string) base.TransportConn {
	var cipher *shadowsocks.Cipher

	cipher, err := shadowsocks.NewCipher(transport.cipherName, transport.password)
	if err != nil {
		log.Fatal("Failed generating ciphers:", err)
	}

	conn, dialErr := shadowsocks.Dial("0.0.0.0:0", address, cipher)
	if dialErr != nil {
		return nil
	}

	transportConn, err := newShadowClientConn(conn)
	if err != nil {
		conn.Close()
		return nil
	}

	return transportConn
}

// Create listener for incoming transport connection
func (transport *shadowTransport) Listen(address string) base.TransportListener {
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

	return newShadowTransportListener(ln, transport)
}

// Methods that implement the base.TransportConn interface
func (transportConn *shadowConn) NetworkConn() net.Conn {
	// This returns the real net.Conn used by the shadowsocks.Conn wrapped by the shadowConn.
	// This may seem confusing, but this is the correct behavior for the semantics
	// required by the PT 2.0 specification.
	// The reason we must wrap it this way is that Go does not allow extension of
	// types defined in another module. So NetworkConn() cannot be defined directly
	// on shadowsocks.Conn.
	return transportConn.conn.Conn
}

// Methods that implement the base.TransportListener interface
// Listener for underlying network connection
func (listener *shadowTransportListener) NetworkListener() net.Listener {
	return listener.listener
}

// Accept waits for and returns the next connection to the listener.
func (listener *shadowTransportListener) TransportAccept() (base.TransportConn, error) {
	conn, err := listener.listener.Accept()
	if err != nil {
		return nil, err
	}

	cipher, err := shadowsocks.NewCipher(listener.transport.cipherName, listener.transport.password)
	if err != nil {
		log.Fatal("Failed generating ciphers:", err)
	}

	ssconn := shadowsocks.NewConn(conn, cipher)

	return newShadowServerConn(ssconn)
}

// Close closes the transport listener.
// Any blocked TransportAccept operations will be unblocked and return errors.
func (listener *shadowTransportListener) Close() error {
	return listener.listener.Close()
}

type shadowConn struct {
	conn *shadowsocks.Conn
}

func (conn *shadowConn) Read(b []byte) (int, error) {
	return conn.Read(b)
}

func (conn *shadowConn) Write(b []byte) (int, error) {
	return conn.Write(b)
}

func (conn *shadowConn) Close() error {
	return conn.Close()
}

func (conn *shadowConn) LocalAddr() net.Addr {
	return conn.LocalAddr()
}

func (conn *shadowConn) RemoteAddr() net.Addr {
	return conn.RemoteAddr()
}

func (conn *shadowConn) SetDeadline(t time.Time) error {
	return conn.SetDeadline(t)
}

func (conn *shadowConn) SetReadDeadline(t time.Time) error {
	return conn.SetReadDeadline(t)
}

func (conn *shadowConn) SetWriteDeadline(t time.Time) error {
	return conn.SetWriteDeadline(t)
}

func newShadowClientConn(conn *shadowsocks.Conn) (c *shadowConn, err error) {
	// Initialize a client connection.
	c = &shadowConn{conn}

	return
}

func newShadowServerConn(conn *shadowsocks.Conn) (c *shadowConn, err error) {
	// Initialize a server connection.
	c = &shadowConn{conn}

	return
}

var _ base.Transport = (*shadowTransport)(nil)
var _ base.TransportListener = (*shadowTransportListener)(nil)
var _ base.TransportConn = (*shadowConn)(nil)
var _ net.Conn = (*shadowConn)(nil)
