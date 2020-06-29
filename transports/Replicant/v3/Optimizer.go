package replicant

import (
	pt "github.com/OperatorFoundation/shapeshifter-ipc"
	"golang.org/x/net/proxy"
	"net"
)

// This makes Replicant compliant with Optimizer
type Transport struct {
	Config  ClientConfig
	Sconfig ServerConfig
	//TODO was adding Sconfig the right move?  after running the tests, nothing in the initial code was broken or failing
	//TODO Do we need to add a test for the dial and listen that take nothing?
	Address string
	Dialer  proxy.Dialer
}

func (transport Transport) Dial() (net.Conn, error) {
	conn, dialErr := transport.Dialer.Dial("tcp", transport.Address)
	if dialErr != nil {
		return nil, dialErr
	}

	dialConn := conn
	transportConn, err := NewClientConnection(conn, transport.Config)
	if err != nil {
		_ = dialConn.Close()
		return nil, err
	}

	return transportConn, nil
	}

func (transport Transport) Listen() (net.Listener, error) {
	addr, resolveErr := pt.ResolveAddr(transport.Address)
	if resolveErr != nil {
		return nil, resolveErr
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return newReplicantTransportListener(ln, transport.Sconfig), nil
}
	//replicantTransport := New(transport.Config, transport.Dialer)
	//conn := replicantTransport.Dial(transport.Address)
	//conn, err:= replicantTransport.Dial(transport.Address), errors.New("connection failed")
	//if err != nil {
	//	return nil, err
	//} else {
	//	return conn, nil
	//}
	//return conn, nil


