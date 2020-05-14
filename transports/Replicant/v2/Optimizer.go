package replicant

import (
	"golang.org/x/net/proxy"
	"net"
)

// This makes Replicant compliant with Optimizer
type Transport struct {
	Config  ClientConfig
	Address string
	Dialer  proxy.Dialer
}

// TODO: the dial we call currently does not return an error
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

	//replicantTransport := New(transport.Config, transport.Dialer)
	//conn := replicantTransport.Dial(transport.Address)
	//conn, err:= replicantTransport.Dial(transport.Address), errors.New("connection failed")
	//if err != nil {
	//	return nil, err
	//} else {
	//	return conn, nil
	//}
	//return conn, nil
}

