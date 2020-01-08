package replicant

import (
	"golang.org/x/net/proxy"
	"net"
)

// This makes Replicant compliant with Optimizer
type Transport struct {
	Config  Config
	Address string
	Dialer  proxy.Dialer
}

// TODO: the dial we call currently does not return an error
func (transport Transport) Dial() (net.Conn, error) {
	replicantTransport:= New(transport.Config, transport.Dialer)
	conn := replicantTransport.Dial(transport.Address)
	//conn, err:= replicantTransport.Dial(transport.Address), errors.New("connection failed")
	//if err != nil {
	//	return nil, err
	//} else {
	//	return conn, nil
	//}
	return conn, nil
}

