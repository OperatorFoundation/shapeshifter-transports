package replicant

import (
	"errors"
	"net"
)

type Transport struct {
	Config  Config
	Address string
}

func (transport Transport) Dial() (net.Conn, error) {
	replicantTransport:= New(transport.Config)
	conn, err:= replicantTransport.Dial(transport.Address), errors.New("connection failed")
	if err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

