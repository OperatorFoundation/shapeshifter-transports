/*
 * Copyright (c) 2019, Operator Foundation
 *
 */

// Package Optimizer provides a PT 2.0 Go API wrapper around the connections used
package Optimizer

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow"
	"net"
)
type Transport interface {
	Dial() net.Conn
}

// optimizerTransport is the optimizer implementation of the base.Transport interface.
type optimizerTransport struct {
	transports []Transport
}

type ShadowTransport struct {
	password   string
	cipherName string
	address string
}

func NewOptimizerTransport(transports []Transport) *optimizerTransport {
	return &optimizerTransport{transports}
}

func NewOptimizerClient (transports []Transport) *optimizerTransport {
	return &optimizerTransport{transports}
}

// Create outgoing transport connection
func (opTransport *optimizerTransport) Dial() net.Conn {
	var transport Transport
	transport = opTransport.transports[0]

	conn := transport.Dial()
	if conn == nil {
		return nil
	}

	return conn
}

func (transport ShadowTransport) Dial() net.Conn {
	shadowTransport := shadow.NewShadowTransport(transport.password, transport.cipherName)
	conn := shadowTransport.Dial(transport.address)
	return conn
}
