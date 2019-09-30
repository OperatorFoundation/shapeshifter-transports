/*
 * Copyright (c) 2019, Operator Foundation
 *
 */

// Package Optimizer provides a PT 2.0 Go API wrapper around the connections used
package Optimizer

import (
	"errors"
	_ "math/rand"
	"net"
	"time"
)

const timeoutInSeconds = 60

type Transport interface {
	Dial() (net.Conn, error)
}

type Client struct {
	Transports []Transport
	Strategy   Strategy
}

func NewOptimizerClient(Transports []Transport, Strategy Strategy) *Client {
	return &Client{Transports, Strategy}
}

func (OptT *Client) Dial() (net.Conn, error) {
	firstTryTime := time.Now()
	transport := OptT.Strategy.Choose()
	if transport == nil {
		return nil, errors.New("optimizer strategy returned nil")
	}
	conn, err := transport.Dial()
	for err != nil {
		OptT.Strategy.Report(transport, false, 60)
		currentTryTime := time.Now()
		durationElapsed := currentTryTime.Sub(firstTryTime)
		if durationElapsed >= timeoutInSeconds {
			return nil, errors.New("timeout. Dial time exceeded")
		}
		transport = OptT.Strategy.Choose()
		conn, err = transport.Dial()
	}
	OptT.Strategy.Report(transport, true, 60)
	return conn, nil
}

type Strategy interface {
	Choose() Transport
	Report(transport Transport, success bool, durationElapsed float64)
}

//begin refactor
type FirstStrategy struct {
	transports []Transport
}

func NewFirstStrategy(transports []Transport) Strategy {
	return &FirstStrategy{transports}
}

func (strategy *FirstStrategy) Choose() Transport {
	return strategy.transports[0]
}

func (strategy *FirstStrategy) Report(transport Transport, success bool, durationElapsed float64) {

}

//end refactor
type RandomStrategy struct {
	transports []Transport
}

func NewRandomStrategy(transports []Transport) Strategy {
	return &RandomStrategy{transports}
}

func (strategy *RandomStrategy) Choose() Transport {
	return strategy.transports[0]
}

func (strategy *RandomStrategy) Report(transport Transport, success bool, durationElapsed float64) {

}

type RotateStrategy struct {
	transports []Transport
	index      int
}

func NewRotateStrategy(transports []Transport) Strategy {
	return &RotateStrategy{transports, 0}
}

func (strategy *RotateStrategy) Choose() Transport {
	transport := strategy.transports[strategy.index]
	strategy.index += 1
	if strategy.index >= len(strategy.transports) {
		strategy.index = 0
	}
	return transport
}

func (strategy *RotateStrategy) Report(transport Transport, success bool, durationElapsed float64) {

}

type TrackStrategy struct {
	index     int
	trackMap  map[Transport]int
	transport []Transport
}

func NewTrackStrategy(transport []Transport) Strategy {
	track := TrackStrategy{0, map[Transport]int{}, transport}
	track.trackMap = make(map[Transport]int)
	return &track
}

func (strategy *TrackStrategy) Choose() Transport {
	transport := strategy.transport[strategy.index]
	score := strategy.findScore(strategy.transport)
	startIndex := strategy.index
	strategy.incrementIndex(strategy.transport)
	for startIndex != strategy.index {
		if score == 1 {
			return transport
		} else {
			transport = strategy.transport[strategy.index]
			score = strategy.findScore(strategy.transport)
			strategy.incrementIndex(strategy.transport)
		}
	}
	return nil
}

func (strategy *TrackStrategy) findScore(transports []Transport) int {
	transport := transports[strategy.index]
	score, ok := strategy.trackMap[transport]
	if ok {
		return score
	} else {
		return 1
	}
}

func (strategy *TrackStrategy) incrementIndex(transports []Transport) {
	strategy.index += 1
	if strategy.index >= len(transports) {
		strategy.index = 0
	}
}

func (strategy *TrackStrategy) Report(transport Transport, success bool, durationElapsed float64) {
	if success {
		strategy.trackMap[transport] = 1
	} else {
		strategy.trackMap[transport] = 0
	}
}

type minimizeDialDuration struct {
	index      int
	trackMap   map[Transport]float64
	transports []Transport
}

func NewMinimizeDialDuration(transport []Transport) Strategy {
	duration := minimizeDialDuration{0, map[Transport]float64{}, transport}
	duration.trackMap = make(map[Transport]float64)
	return &duration
}

func (strategy *minimizeDialDuration) Choose() Transport {
	transport := strategy.transports[strategy.index]
	score := strategy.findScore(strategy.transports)
	startIndex := strategy.index
	strategy.incrementIndex(strategy.transports)
	for startIndex != strategy.index {
		if score == 0 {
			return transport
		} else {
			strategy.incrementIndex(strategy.transports)
			transport = strategy.minDuration()
			if transport == nil {
				transport = strategy.transports[strategy.index]
				score = strategy.findScore(strategy.transports)
				continue
			} else {
				return transport
			}
		}
	}
	return nil
}

func (strategy *minimizeDialDuration) incrementIndex(transports []Transport) {
	strategy.index += 1
	if strategy.index >= len(transports) {
		strategy.index = 0
	}
}

func (strategy *minimizeDialDuration) findScore(transports []Transport) float64 {
	transport := transports[strategy.index]
	score, ok := strategy.trackMap[transport]
	if ok {
		return score
	} else {
		return 0
	}
}

func (strategy *minimizeDialDuration) Report(transport Transport, success bool, durationElapsed float64) {
	if success {
		if durationElapsed < 60 {
			strategy.trackMap[transport] = durationElapsed
		} else {
			strategy.trackMap[transport] = 60.0
		}
	} else {
		strategy.trackMap[transport] = 60.0
	}
}

func (strategy *minimizeDialDuration) minDuration() Transport {
	min := 61.0
	var transport Transport = nil
	for key, value := range strategy.trackMap {
		if value < min {
			min = value
			transport = key
		}
	}
	return transport
}
