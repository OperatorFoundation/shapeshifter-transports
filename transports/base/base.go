/*
 * Copyright (c) 2014, Yawning Angel <yawning at torproject dot org>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *  * Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 *
 *  * Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

// Package base provides the common interface that each supported transport
// protocol must implement.
package base

import (
	"net"

	"github.com/OperatorFoundation/shapeshifter-ipc"
)

// Pluggable Transport Specification v2.0, draft 1
// 3.2.4.1.1. Module pt
// The Transport interface provides a way to make outgoing transport connections and to accept
// incoming transport connections.
// It also exposes access to an underlying network connection Dialer.
// The Dialer can be modified to change how the network connections are made.
interface Transport {
	// Dialer for the underlying network connection
	networkDialer() *Dialer

	// Create outgoing transport connection
	(transport *Transport) Dial(address string) pt.TransportConn

	// Create listener for incoming transport connection
	(transport *Transport) Listen(address string) pt.TransportListener
}

// The TransportConn interface represents a transport connection.
// The primary function of a transport connection is to provide the net.Conn interface.
// This interface also exposes access to an underlying network connection,
// which also implements net.Conn.
interface TransportConn extends net.Conn {
	// Conn for the underlying network connection
	networkConn *Conn
}

// The TransportListener interface represents a listener for a transport connection.
// This interface also exposes access to an underlying network listener.
interface TransportListener {
	// Listener for underlying network connection
	networkListener *Listener

	// Accept waits for and returns the next connection to the listener.
	TransportAccept() (TransportConn, error)

	// Close closes the transport listener.
	// Any blocked TransportAccept operations will be unblocked and return errors.
	Close() error
}
