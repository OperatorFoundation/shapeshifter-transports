package replicant

import (
	"errors"
	"golang.org/x/net/proxy"
	"net"
)

// This makes Replicant compliant with Optimizer
type Transport struct {
	Config  Config
	Address string
	Dialer  proxy.Dialer
}

func (transport Transport) Dial() (net.Conn, error) {
	replicantTransport:= New(transport.Config, transport.Dialer)
	conn, err:= replicantTransport.Dial(transport.Address), errors.New("connection failed")
	if err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

