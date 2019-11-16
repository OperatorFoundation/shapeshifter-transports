/*
 * Copyright (c) 2019, Operator Foundation
 *
 */

// Package replicant provides a PT 2.1 Go API implementation of the Replicant adversary-tunable transport
package replicant

import (
	"bytes"
	"fmt"
	"net"

	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/toneburst"
)

// replicantTransport is the replicant implementation of the base.Transport interface.
type replicantTransport struct {
	config 		Config
}

type ReplicantConnectionState struct {
	toneburst toneburst.ToneBurst
	polish    polish.PolishConnection
}

type ReplicantConnection struct {
	state *ReplicantConnectionState
	conn net.Conn
	receiveBuffer *bytes.Buffer
}

type ReplicantServer struct {
	toneburst toneburst.ToneBurst
	polish    polish.PolishServer
}

type replicantTransportListener struct {
	listener  *net.TCPListener
	transport *replicantTransport
}

func New(config Config) *replicantTransport {
	return &replicantTransport{config: config}
}

func newReplicantTransportListener(listener *net.TCPListener, transport *replicantTransport) *replicantTransportListener {
	return &replicantTransportListener{listener: listener, transport: transport}
}

func NewClientConnection(conn net.Conn, config Config) (*ReplicantConnection, error) {
	// Initialize a client connection.
	var buffer bytes.Buffer

	state := NewReplicantClientConnectionState(config)
	rconn := &ReplicantConnection{state, conn, &buffer}

	err := state.toneburst.Perform(conn)
	if err != nil {
		fmt.Println("Toneburst failed")
		return nil, err
	}

	err = state.polish.Handshake(conn)
	if err != nil {
		fmt.Println("Polish handshake failed")
		return nil, err
	}

	return rconn, nil
}

func NewServerConnection(conn net.Conn, config Config) (*ReplicantConnection, error) {
	// Initialize a client connection.
	var buffer bytes.Buffer

	state := NewReplicantClientConnectionState(config)
	rconn := &ReplicantConnection{state, conn, &buffer}

	err := state.toneburst.Perform(conn)
	if err == nil {
		return nil, err
	}

	err = state.polish.Handshake(conn)
	if err != nil {
		fmt.Println("Polish handshake failed")
		return nil, err
	}

	return rconn, nil
}

func NewReplicantClientConnectionState(config Config) *ReplicantConnectionState {
	toneburst := toneburst.New(config.Toneburst)
	polish := polish.NewClient(config.Polish)

	return &ReplicantConnectionState{toneburst, polish}
}

func NewReplicantServerConnectionState(config Config, polishServer polish.PolishServer, conn net.Conn) *ReplicantConnectionState {
	toneburst := toneburst.New(config.Toneburst)
	polish := polishServer.NewConnection(conn)

	return &ReplicantConnectionState{toneburst, polish}
}

