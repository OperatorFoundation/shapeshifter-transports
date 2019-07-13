/*
 * Copyright (c) 2019, Operator Foundation
 *
 */

// Package Optimizer provides a PT 2.0 Go API wrapper around the connections used
package Optimizer

import (
	"golang.org/x/net/proxy"
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

func NewOptimizerClient(transports []Transport, strategy Strategy) *optimizerTransport {
	return &optimizerTransport{transports, strategy}
}

//func (opTransport *optimizerTransport) Dial() net.Conn {
//	transport := opTransport.strategy.Choose(opTransport.transports)
//
//	conn:= transport.Dial()
//	if conn == nil {
//		return nil
//	}
//
//	return conn
//}
func (opTransport *optimizerTransport) Dial(address string) net.Conn {
	dialFn := proxy.Direct.Dial
	transport := opTransport.strategy.Choose(opTransport.transports)

	conn, dialErr := dialFn("tcp", address)
	if dialErr != nil {
		//find a way to move to the next or pass the current one
		return nil
	}

	conn = transport.Dial()
	if conn == nil {
		return nil
	}

	return conn
}

type Strategy interface {
	Choose([]Transport) Transport
	Report(transport Transport, success bool)
}

type FirstStrategy struct {

}

func (strategy FirstStrategy) Choose(transports []Transport) Transport {
	return transports[0]
}

func (strategy FirstStrategy) Report(transport Transport, success bool) {

}

type RandomStrategy struct {

}

func (strategy RandomStrategy) Choose(transports []Transport) Transport {
	return transports[rand.Intn(len(transports))]
}

func (strategy RandomStrategy) Report(transport Transport, success bool) {

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

func (strategy RotateStrategy) Report(transport Transport, success bool) {

}

type TrackStrategy struct {
index int
}

func (strategy TrackStrategy) Choose(transports []Transport) Transport {
	//i think the track/feedback strat is going to need to cycle through transports like rotate does
	transport := transports[strategy.index]
	strategy.index += 1
	if strategy.index >= len(transports) {
		strategy.index = 0
	}
	return transport
}

func (strategy TrackStrategy) Report(transport Transport, success bool) {
	map[Transport]float64
	//if _  == <true>
	//_
	}
