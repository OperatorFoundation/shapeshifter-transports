/*
	MIT License

	Copyright (c) 2020 Operator Foundation

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

// Package optimizer provides a PT 2.0 Go API wrapper around the connections used
package optimizer

import (
	"errors"
	"net"
	"time"
)

const timeoutInSeconds = 60

//TransportDialer is a program that transforms network traffic
type TransportDialer interface {
	Dial() (net.Conn, error)
}

type TransportListener interface {
	Listen() (net.Listener, error)
}

//Client contains the two parameters needed to use Optimizer.
type Client struct {
	Transports []TransportDialer
	Strategy   Strategy
}

//NewOptimizerClient is the initializer
func NewOptimizerClient(Transports []TransportDialer, Strategy Strategy) *Client {
	return &Client{Transports, Strategy}
}

//Dial connects to the address on the named network
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

//Strategy is the method used to choose a transport
type Strategy interface {
	Choose() TransportDialer
	Report(transport TransportDialer, success bool, durationElapsed float64)
}

//FirstStrategy returns the first strategy in the array
type FirstStrategy struct {
	transports []TransportDialer
}

//NewFirstStrategy initializes FirstStrategy
func NewFirstStrategy(transports []TransportDialer) Strategy {
	return &FirstStrategy{transports}
}

//Choose selects a transport in the array
func (strategy *FirstStrategy) Choose() TransportDialer {
	return strategy.transports[0]
}

//Report returns if the transport was successful and how long the connection took
func (strategy *FirstStrategy) Report(TransportDialer, bool, float64) {

}

//NewRandomStrategy initializes RandomStrategy
func NewRandomStrategy(transports []TransportDialer) Strategy {
	return &RandomStrategy{transports}
}

//RandomStrategy returns a transport at random
type RandomStrategy struct {
	transports []TransportDialer
}

//Choose selects a transport in the array
func (strategy *RandomStrategy) Choose() TransportDialer {
	return strategy.transports[0]
}

//Report returns if the transport was successful and how long the connection took
func (strategy *RandomStrategy) Report(TransportDialer, bool, float64) {

}

//NewRotateStrategy initializes RotateStrategy
func NewRotateStrategy(transports []TransportDialer) Strategy {
	return &RotateStrategy{transports, 1}
}

//RotateStrategy cycles through the list of transports, using a different one each time
type RotateStrategy struct {
	transports []TransportDialer
	index      int
}

//Choose selects a transport in the array
func (strategy *RotateStrategy) Choose() TransportDialer {
	transport := strategy.transports[strategy.index]
	strategy.index++
	if strategy.index >= len(strategy.transports) {
		strategy.index = 0
	}
	return transport
}

//Report returns if the transport was successful and how long the connection took
func (strategy *RotateStrategy) Report(TransportDialer, bool, float64) {

}

//TrackStrategy assigns a score to each transport and server that is remembered and used to
//choose the best option
type TrackStrategy struct {
	index     int
	trackMap  map[TransportDialer]int
	transport []TransportDialer
}

//NewTrackStrategy initializes TrackStrategy
func NewTrackStrategy(transport []TransportDialer) Strategy {
	track := TrackStrategy{0, map[TransportDialer]int{}, transport}
	track.trackMap = make(map[TransportDialer]int)
	return &track
}

//Choose selects a transport in the array
func (strategy *TrackStrategy) Choose() TransportDialer {
	transport := strategy.transport[strategy.index]
	score := strategy.findScore(strategy.transport)
	startIndex := strategy.index
	strategy.incrementIndex(strategy.transport)
	for startIndex != strategy.index {
		if score == 1 {
			return transport
		}
		transport = strategy.transport[strategy.index]
		score = strategy.findScore(strategy.transport)
		strategy.incrementIndex(strategy.transport)

	}
	return nil
}

//findScore is used to find the score given to each transport based on performance
func (strategy *TrackStrategy) findScore(transports []TransportDialer) int {
	transport := transports[strategy.index]
	score, ok := strategy.trackMap[transport]
	if ok {
		return score
	}
	return 1

}

//incrementIndex is used to cycle through the transport index
func (strategy *TrackStrategy) incrementIndex(transports []TransportDialer) {
	strategy.index++
	if strategy.index >= len(transports) {
		strategy.index = 0
	}
}

//Report returns if the transport was successful and how long the connection took
func (strategy *TrackStrategy) Report(transport TransportDialer, success bool, _ float64) {
	if success {
		strategy.trackMap[transport] = 1
	} else {
		strategy.trackMap[transport] = 0
	}
}

//minimizeDialDuration is used to find the transport with the fastest response time
type minimizeDialDuration struct {
	index      int
	trackMap   map[TransportDialer]float64
	transports []TransportDialer
}

//NewMinimizeDialDuration initializes minimizeDialDuration
func NewMinimizeDialDuration(transport []TransportDialer) Strategy {
	duration := minimizeDialDuration{0, map[TransportDialer]float64{}, transport}
	duration.trackMap = make(map[TransportDialer]float64)
	return &duration
}

//Choose selects a transport in the array
func (strategy *minimizeDialDuration) Choose() TransportDialer {
	transport := strategy.transports[strategy.index]
	score := strategy.findScore(strategy.transports)
	startIndex := strategy.index
	strategy.incrementIndex(strategy.transports)
	for startIndex != strategy.index {
		if score == 0 {
			return transport
		}
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
	return nil
}

//incrementIndex is used to cycle through the transport index
func (strategy *minimizeDialDuration) incrementIndex(transports []TransportDialer) {
	strategy.index++
	if strategy.index >= len(transports) {
		strategy.index = 0
	}
}

//findScore is used to find the score given to each transport based on performance
func (strategy *minimizeDialDuration) findScore(transports []TransportDialer) float64 {
	transport := transports[strategy.index]
	score, ok := strategy.trackMap[transport]
	if ok {
		return score
	}
	return 0

}

//Report returns if the transport was successful and how long the connection took
func (strategy *minimizeDialDuration) Report(transport TransportDialer, success bool, durationElapsed float64) {
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

//minDuration assigns a value to the response time
func (strategy *minimizeDialDuration) minDuration() TransportDialer {
	min := 61.0
	var transport TransportDialer = nil
	for key, value := range strategy.trackMap {
		if value < min {
			min = value
			transport = key
		}
	}
	return transport
}
