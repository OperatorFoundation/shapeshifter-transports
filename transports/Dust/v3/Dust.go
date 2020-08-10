package Dust

/*
	MIT License

	Copyright (c) 2020 Operator Foundation

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

import (
	"fmt"
	"github.com/kataras/golog"
	"golang.org/x/net/proxy"
	"net"
	"time"

	"github.com/OperatorFoundation/shapeshifter-ipc/v2"
	"github.com/blanu/Dust/go/v2/interface"
)

type dustClient struct {
	serverPubkey *Dust.ServerPublic
	dialer       proxy.Dialer
	Address      string `json:"address"`
}

type dustServer struct {
	serverPrivkey *Dust.ServerPrivate
	dialer        proxy.Dialer
}

func NewDustClient(serverPublic string, dialer proxy.Dialer, address string) *dustClient {
	unparsed := make(map[string]string)
	unparsed["p"] = serverPublic

	spub, err := Dust.ParseServerPublic(unparsed)
	if err != nil {
		return nil
	}

	return &dustClient{serverPubkey: spub, dialer: dialer, Address: address}
}

type dustTransportListener struct {
	listener  *net.TCPListener
	transport *dustServer
}

//begin optimizer code
type Transport struct {
	ServerPublic string
	Address      string
	Dialer       proxy.Dialer
	ServerConfig *dustServer
}

type Config struct {
	ServerPublic string `json:"server-public"`
	Address      string `json:"address"`
}

func New(serverPublic string, address string, dialer proxy.Dialer, serverConfig *dustServer) Transport {
	return Transport{
		ServerPublic: serverPublic,
		Address:      address,
		Dialer:       dialer,
		ServerConfig: serverConfig,
	}
}

func (transport Transport) Dial() (net.Conn, error) {
	dustTransport := NewDustClient(transport.ServerPublic, transport.Dialer, transport.Address)
	conn, err := dustTransport.Dial(transport.Address)
	if err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

func (transport Transport) Listen() (net.Listener, error) {
	addr, resolveErr := pt.ResolveAddr(transport.Address)
	if resolveErr != nil {
		fmt.Println(resolveErr.Error())
		return nil, resolveErr
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return newDustTransportListener(ln, transport.ServerConfig), nil
}

//end optimizer code

func newDustTransportListener(listener *net.TCPListener, transport *dustServer) *dustTransportListener {
	return &dustTransportListener{listener: listener, transport: transport}
}

// Create outgoing transport connection
func (transport *dustClient) Dial(address string) (net.Conn, error) {
	conn, dialErr := net.Dial("tcp", address)
	if dialErr != nil {
		return conn, dialErr
	}

	transportConn, err := Dust.BeginRawStreamClient(conn, transport.serverPubkey)
	if err != nil {
		closeErr := conn.Close()
		if closeErr != nil {
			golog.Error("could not close")
		}
		return conn, dialErr
	}

	return transportConn, err
}

// Create listener for incoming transport connection
func (transport *dustServer) Listen(address string) (net.Listener, error) {
	addr, resolveErr := pt.ResolveAddr(address)
	if resolveErr != nil {
		fmt.Println(resolveErr.Error())
		return nil, resolveErr
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return newDustTransportListener(ln, transport), nil
}

func (listener *dustTransportListener) Addr() net.Addr {
	interfaces, _ := net.Interfaces()
	addrs, _ := interfaces[0].Addrs()
	return addrs[0]
}

// Accept waits for and returns the next connection to the listener.
func (listener *dustTransportListener) Accept() (net.Conn, error) {
	conn, err := listener.listener.Accept()
	if err != nil {
		return nil, err
	}

	ssconn, streamerr := Dust.BeginRawStreamServer(conn, listener.transport.serverPrivkey)
	if streamerr != nil {
		return nil, streamerr
	}

	dsconn, dserr := newDustServerConn(ssconn)
	if dserr != nil {
		return nil, dserr
	}

	return dsconn, nil
}

// Close closes the transport listener.
// Any blocked Accept operations will be unblocked and return errors.
func (listener *dustTransportListener) Close() error {
	return listener.listener.Close()
}

type dustConn struct {
	conn *Dust.RawStreamConn
}

func (sconn *dustConn) Read(b []byte) (int, error) {
	return sconn.conn.Read(b)
}

func (sconn *dustConn) Write(b []byte) (int, error) {
	return sconn.conn.Write(b)
}

func (sconn *dustConn) Close() error {
	return sconn.conn.Close()
}

func (sconn *dustConn) LocalAddr() net.Addr {
	return sconn.conn.LocalAddr()
}

func (sconn *dustConn) RemoteAddr() net.Addr {
	return sconn.conn.RemoteAddr()
}

func (sconn *dustConn) SetDeadline(t time.Time) error {
	return sconn.conn.SetDeadline(t)
}

func (sconn *dustConn) SetReadDeadline(t time.Time) error {
	return sconn.conn.SetReadDeadline(t)
}

func (sconn *dustConn) SetWriteDeadline(t time.Time) error {
	return sconn.conn.SetWriteDeadline(t)
}

func newDustServerConn(conn *Dust.RawStreamConn) (c *dustConn, err error) {
	// Initialize a server connection.
	c = &dustConn{conn}

	return
}

var _ net.Listener = (*dustTransportListener)(nil)
var _ net.Conn = (*dustConn)(nil)
