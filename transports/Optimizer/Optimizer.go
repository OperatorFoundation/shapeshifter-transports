/*
 * Copyright (c) 2019, Operator Foundation
 *
 */

// Package Optimizer provides a PT 2.0 Go API wrapper around the connections used
package Optimizer

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow"
	_ "github.com/OperatorFoundation/shapeshifter-transports/transports/shadow"
	"math/rand"
	"net"
)

type Transport interface {
	Dial() net.Conn
}

type optimizerTransport struct {
	transports []Transport
	strategy   Strategy
}

type ShadowTransport struct {
	password   string
	cipherName string
	address    string
}

type Obfs4Transport struct {
	certString string
	iatMode    int
	address    string
}

func NewOptimizerClient(transports []Transport, strategy Strategy) *optimizerTransport {
	return &optimizerTransport{transports, strategy}
}

func (opTransport *optimizerTransport) Dial() net.Conn {
	transport := opTransport.strategy.Choose(opTransport.transports)

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
	conn := Obfs4Transport.Dial(transport.address)
	return conn
}

type Strategy interface {
	Choose([]Transport) Transport
}

type FirstStrategy struct {
}

func (strategy FirstStrategy) Choose(transports []Transport) Transport {
	return transports[0]
}

type RandomStrategy struct {
}

func (strategy RandomStrategy) Choose(transports []Transport) Transport {
	return transports[rand.Intn(len(transports))]
}

type RotateStrategy struct {
	index int
}

func (strategy RotateStrategy) Choose(transports []Transport) Transport {
	transport := transports[strategy.index]
	strategy.index += 1
	if strategy.index >= len(transports) {
		strategy.index = 0
	}
	return transport
}
