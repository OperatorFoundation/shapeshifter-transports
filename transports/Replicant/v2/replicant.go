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

	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2/toneburst"
)

type ConnectionState struct {
	toneburst toneburst.ToneBurst
	polish    polish.Connection
}

type Connection struct {
	state *ConnectionState
	conn net.Conn
	receiveBuffer *bytes.Buffer
}

type Server struct {
	toneburst toneburst.ToneBurst
	polish    polish.Server
}

type replicantTransportListener struct {
	listener  *net.TCPListener
	config ServerConfig
}

func newReplicantTransportListener(listener *net.TCPListener, config ServerConfig) *replicantTransportListener {
	return &replicantTransportListener{listener: listener, config: config}
}

func NewClientConnection(conn net.Conn, config ClientConfig) (*Connection, error) {
	// Initialize a client connection.
	var buffer bytes.Buffer

	state, clientError := NewReplicantClientConnectionState(config)
	if clientError != nil {
		return nil, clientError
	}
	rconn := &Connection{state, conn, &buffer}

	if state.toneburst != nil {
		err := state.toneburst.Perform(conn)
		if err != nil {
			return nil, err
		}

	}
	//FIXME: Handshake when polish is nil
	if state.polish != nil {
		err := state.polish.Handshake(conn)
		if err != nil {
			return nil, err
		}
	}

	return rconn, nil
}

func NewServerConnection(conn net.Conn, config ServerConfig) (*Connection, error) {
	// Initialize a client connection.
	var buffer bytes.Buffer
	var polishServer polish.Server
	var serverError error

	if config.Polish != nil {
		polishServer, serverError = config.Polish.Construct()
		if serverError != nil {
			return nil, serverError
		}
	}

	state, connError := NewReplicantServerConnectionState(config, polishServer, conn)
	if connError != nil {
		return nil, connError
	}
	rconn := &Connection{state, conn, &buffer}

	if state.toneburst != nil {
		err := state.toneburst.Perform(conn)
		if err != nil {
			fmt.Println("> Toneburst error: ", err.Error())
			return nil, err
		}

		println("> Performed toneburst succesfully.")
	}

	if state.polish != nil {
		err := state.polish.Handshake(conn)
		if err != nil {
			fmt.Println("> Polish handshake failed", err.Error())
			return nil, err
		}

		println("> Successful polish handshake.")
	}

	println("> New server connection created.")
	return rconn, nil
}

func NewReplicantClientConnectionState(config ClientConfig) (*ConnectionState, error) {
	var tb toneburst.ToneBurst
	var toneburstError error
	var p polish.Connection
	var polishError error

	if config.Toneburst != nil {
		tb, toneburstError = config.Toneburst.Construct()
		if toneburstError != nil {
			return nil, toneburstError
		}
	}

	if config.Polish != nil {
		p, polishError = config.Polish.Construct()
		if polishError != nil {
			return nil, polishError
		}
	}


	return &ConnectionState{tb, p}, nil
}

func NewReplicantServerConnectionState(config ServerConfig, polishServer polish.Server, conn net.Conn) (*ConnectionState, error) {
	var tb toneburst.ToneBurst
	var toneburstError error
	var p polish.Connection

	if config.Toneburst != nil {
		tb, toneburstError = config.Toneburst.Construct()
		if toneburstError != nil {
			return nil, toneburstError
		}
	}

	if polishServer != nil {
		p = polishServer.NewConnection(conn)
	}

	return &ConnectionState{tb, p}, nil
}
