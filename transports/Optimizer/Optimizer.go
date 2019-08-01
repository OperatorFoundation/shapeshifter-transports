/*
 * Copyright (c) 2019, Operator Foundation
 *
 */

// Package Optimizer provides a PT 2.0 Go API wrapper around the connections used
package Optimizer

import (
	"errors"
	"math/rand"
	"net"
	"time"
)

const timeoutInSeconds = 60

type Transport interface {
	Dial() (net.Conn, error)
}

type optimizerTransport struct {
	transports []Transport
	strategy   Strategy
}

func NewOptimizerClient(transports []Transport, strategy Strategy) *optimizerTransport {
	return &optimizerTransport{transports, strategy}
}

func (opTransport *optimizerTransport) Dial() (net.Conn, error) {
	firstTryTime := time.Now()
	transport := opTransport.strategy.Choose(opTransport.transports)
	if transport == nil {
		return nil, errors.New("optimizer strategy returned nil")
	}
	conn, err := transport.Dial()
	for err != nil {
		opTransport.strategy.Report(transport, false, 60)
		currentTryTime := time.Now()
		durationElapsed := currentTryTime.Sub(firstTryTime)
		if durationElapsed >= timeoutInSeconds {
			return nil, errors.New("timeout. Dial time exceeded")
		}
		transport = opTransport.strategy.Choose(opTransport.transports)
		conn, err = transport.Dial()
	}
	opTransport.strategy.Report(transport, true, 60)
	return conn, nil
}

type Strategy interface {
	Choose([]Transport) Transport
	Report(transport Transport, success bool, durationElapsed float64)
}

type FirstStrategy struct {
}

func NewFirstStrategy() Strategy {
	return FirstStrategy{}
}

func (strategy FirstStrategy) Choose(transports []Transport) Transport {
	return transports[0]
}

func (strategy FirstStrategy) Report(transport Transport, success bool, durationElapsed float64) {

}

type RandomStrategy struct {
}

func NewRandomStrategy() Strategy{
	return RandomStrategy{}
}

func (strategy RandomStrategy) Choose(transports []Transport) Transport {
	return transports[rand.Intn(len(transports))]
}

func (strategy RandomStrategy) Report(transport Transport, success bool, durationElapsed float64) {

}

type RotateStrategy struct {
	index int
}

func NewRotateStrategy() Strategy{
	return RotateStrategy{}
}

func (strategy RotateStrategy) Choose(transports []Transport) Transport {
	transport := transports[strategy.index]
	strategy.index += 1
	if strategy.index >= len(transports) {
		strategy.index = 0
	}
	return transport
}

func (strategy RotateStrategy) Report(transport Transport, success bool, durationElapsed float64) {

}

type TrackStrategy struct {
	index    int
	trackMap map[Transport]int
}

func NewTrackStrategy() *TrackStrategy {
	track := TrackStrategy{}
	track.trackMap = make(map[Transport]int)
	return &track
}

func (strategy *TrackStrategy) Choose(transports []Transport) Transport {
	transport := transports[strategy.index]
	score := strategy.findScore(transports)
	startIndex := strategy.index
	strategy.incrementIndex(transports)
	for startIndex != strategy.index {
		if score == 1 {
			return transport
		} else {
			transport = transports[strategy.index]
			score = strategy.findScore(transports)
			strategy.incrementIndex(transports)
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
	index    int
	trackMap map[Transport]float64
}

func NewMinimizeDialDuration() *minimizeDialDuration {
	duration := minimizeDialDuration{}
	duration.trackMap = make(map[Transport]float64)
	return &duration
}

func (strategy *minimizeDialDuration) Choose(transports []Transport) Transport {
	transport := transports[strategy.index]
	score := strategy.findScore(transports)
	startIndex := strategy.index
	strategy.incrementIndex(transports)
	for startIndex != strategy.index {
		if score == 0 {
			return transport
		} else {
			strategy.incrementIndex(transports)
			transport = strategy.minDuration()
			if transport == nil {
				transport = transports[strategy.index]
				score = strategy.findScore(transports)
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
