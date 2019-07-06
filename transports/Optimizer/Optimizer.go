/*
 * Copyright (c) 2019, Operator Foundation
 *
 */

// Package Optimizer provides a PT 2.0 Go API wrapper around the connections used
package Optimizer

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow"
	 _"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow"
	"net"
)
type Transport interface {
	Dial() net.Conn
}

type optimizerTransport struct {
	transports []Transport
}

type ShadowTransport struct {
	password   string
	cipherName string
	address    string
}

type Obfs4Transport struct {
	certString string
	iatMode    int
}

func NewOptimizerTransport(transports []Transport) *optimizerTransport {
	return &optimizerTransport{transports}
}

func NewOptimizerClient (transports []Transport) *optimizerTransport {
	return &optimizerTransport{transports}
}

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

func (transport Obfs4Transport) Dial() net.Conn {
	Obfs4Transport := obfs4.NewObfs4Client(transport.certString, transport.iatMode)
	conn := Obfs4Transport.Dial(transport.certString)
	return conn
}