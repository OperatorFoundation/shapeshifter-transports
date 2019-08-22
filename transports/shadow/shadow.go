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
)

// shadowTransport is the shadow implementation of the base.Transport interface.
type shadowTransport struct {
	dialer *net.Dialer

	password   string
	cipherName string
}

func NewShadowClient(password string, cipherName string) *shadowTransport {
	return &shadowTransport{dialer: nil, password: password, cipherName: cipherName}
}

func NewShadowServer(password string, cipherName string) *shadowTransport {
	return &shadowTransport{dialer: nil, password: password, cipherName: cipherName}
}

type shadowTransportListener struct {
	listener  *net.TCPListener
	transport *shadowTransport
}
//begin code added from optimizer
type Transport struct {
	Password   string
	CipherName string
	Address    string
}

func (transport Transport) Dial() (net.Conn, error) {
	shadowTransport := NewShadowClient(transport.Password, transport.CipherName)
	conn, err := shadowTransport.Dial(transport.Address)
	if err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}
//end code added from optimizer
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
func (transport *shadowTransport) Dial(address string) (net.Conn, error) {
	//could maybe use the idiomatic way to introduce a variable
	var cipher *shadowsocks.Cipher

	cipher, err := shadowsocks.NewCipher(transport.cipherName, transport.password)
	if err != nil {
		log.Fatal("Failed generating ciphers:", err)
	}

	conn, dialErr := shadowsocks.Dial("0.0.0.0:0", address, cipher)
	if dialErr != nil {
		return nil, dialErr
	}

	transportConn, transportErr := newShadowClientConn(conn)
	if transportErr != nil {
		_ = conn.Close()
		return nil, transportErr
	} else {
		return transportConn, nil
	}
}

// Create listener for incoming transport connection
func (transport *shadowTransport) Listen(address string) net.Listener {
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

func (listener *shadowTransportListener) Addr() net.Addr {
	interfaces, _ := net.Interfaces()
	addrs, _ := interfaces[0].Addrs()
	return addrs[0]
}

// Methods that implement the net.Conn interface
func (sconn *shadowConn) NetworkConn() net.Conn {
	// This returns the real net.Conn used by the shadowsocks.Conn wrapped by the shadowConn.
	// This may seem confusing, but this is the correct behavior for the semantics
	// required by the PT 2.0 specification.
	// The reason we must wrap it this way is that Go does not allow extension of
	// types defined in another module. So NetworkConn() cannot be defined directly
	// on shadowsocks.Conn.
	return sconn.conn.Conn
}

// Methods that implement the net.Listener interface
// Listener for underlying network connection
func (listener *shadowTransportListener) NetworkListener() net.Listener {
	return listener.listener
}

// Accept waits for and returns the next connection to the listener.
func (listener *shadowTransportListener) Accept() (net.Conn, error) {
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
// Any blocked Accept operations will be unblocked and return errors.
func (listener *shadowTransportListener) Close() error {
	return listener.listener.Close()
}

type shadowConn struct {
	conn *shadowsocks.Conn
}

func (sconn *shadowConn) Read(b []byte) (int, error) {
	return sconn.conn.Read(b)
}

func (sconn *shadowConn) Write(b []byte) (int, error) {
	return sconn.conn.Write(b)
}

func (sconn *shadowConn) Close() error {
	return sconn.conn.Close()
}

func (sconn *shadowConn) LocalAddr() net.Addr {
	return sconn.conn.LocalAddr()
}

func (sconn *shadowConn) RemoteAddr() net.Addr {
	return sconn.conn.RemoteAddr()
}

func (sconn *shadowConn) SetDeadline(t time.Time) error {
	return sconn.conn.SetDeadline(t)
}

func (sconn *shadowConn) SetReadDeadline(t time.Time) error {
	return sconn.conn.SetReadDeadline(t)
}

func (sconn *shadowConn) SetWriteDeadline(t time.Time) error {
	return sconn.conn.SetWriteDeadline(t)
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

var _ net.Listener = (*shadowTransportListener)(nil)
var _ net.Conn = (*shadowConn)(nil)
