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

package replicant

import (
	pt "github.com/OperatorFoundation/shapeshifter-ipc/v2"
	"net"
	"time"
)

// Create outgoing transport connection
func (config ClientConfig) Dial(address string) net.Conn {
	conn, dialErr := net.Dial("tcp", address)
	if dialErr != nil {
		return nil
	}

	transportConn, err := NewClientConnection(conn, config)
	if err != nil {
		if conn != nil {
			_ = conn.Close()
		}
		return nil
	}

	return transportConn
}

// Create listener for incoming transport connection
func (config ServerConfig) Listen(address string) net.Listener {
	addr, resolveErr := pt.ResolveAddr(address)
	if resolveErr != nil {
		return nil
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil
	}

	return newReplicantTransportListener(ln, config)
}

func (listener *replicantTransportListener) Addr() net.Addr {
	interfaces, _ := net.Interfaces()
	addrs, _ := interfaces[0].Addrs()
	return addrs[0]
}

// Accept waits for and returns the next connection to the listener.
func (listener *replicantTransportListener) Accept() (net.Conn, error) {
	conn, err := listener.listener.Accept()
	if err != nil {
		return nil, err
	}

	config := listener.config
	return NewServerConnection(conn, config)
}

// Close closes the transport listener.
// Any blocked Accept operations will be unblocked and return errors.
func (listener *replicantTransportListener) Close() error {
	return listener.listener.Close()
}

func (sconn *Connection) Read(b []byte) (int, error) {

	if sconn.state.polish != nil {

		if len(b) == 0 {
			return 0, nil
		}

		if sconn.receiveBuffer.Len() > 0 {
			if len(b) <= sconn.receiveBuffer.Len() {
				// Read the decrypted data into the provided slice "b"
				readLen, readError := sconn.receiveBuffer.Read(b)
				if readError != nil {
					return 0, readError
				}

				// We've delivered all unpolished data from the buffer, let's clear it out
				// Read just moves the offset but leaves the data in place, we need to reset the buffer
				if sconn.receiveBuffer.Len() == 0 {
					sconn.receiveBuffer.Reset()
				}

				return readLen, nil
			} else {
				result := make([]byte, len(b))

				// Read the decrypted data into the provided slice "b"
				readLen, readError := sconn.receiveBuffer.Read(result)
				sconn.receiveBuffer.Reset()
				if readError != nil {
					return 0, readError
				}

				stillNeeded := result[readLen:]
				snReadLen, snReadError := sconn.Read(stillNeeded)
				if snReadError != nil {
					return 0, snReadError
				}

				// Keep reading until we get the full requested length, or an error
				for snReadLen < len(stillNeeded) {
					// Keep reading, we didn't get enough data
					stillNeeded = stillNeeded[snReadLen:]
					snReadLen, snReadError = sconn.Read(stillNeeded)
					if snReadError != nil {
						return 0, snReadError
					}
				}

				copy(b, result)
				return len(result), nil
			}
		}

		// Make sure to discard stale data by overwriting b with zeros
		zeros := make([]byte, len(b))
		copy(b, zeros)

		chunkSize := sconn.state.polish.GetChunkSize()
		polished := make([]byte, chunkSize)

		// Read encrypted data from the connection and put it into our polished slice
		connReadLen, connReadError := sconn.conn.Read(polished)
		if connReadError != nil {
			return 0, connReadError
		}

		for connReadLen < chunkSize {
			// Keep reading, we didn't get enough data
			partialPolish := polished[connReadLen:]
			connReadLen, connReadError = sconn.conn.Read(partialPolish)
			if connReadError != nil {
				return 0, connReadError
			}
		}

		// Decrypt the data
		unpolished, unpolishError := sconn.state.polish.Unpolish(polished)
		if unpolishError != nil {
			return 0, unpolishError
		}

		if len(unpolished) <= len(b) {
			copy(b, unpolished)
			return len(unpolished), nil
		} else {
			// This will copy only up to the length of b from unpolished to b
			copy(b, unpolished)

			remainingUnpolished := unpolished[len(b):]

			// Empty the buffer and write the decrypted data to it
			sconn.receiveBuffer.Reset()
			sconn.receiveBuffer.Write(remainingUnpolished)

			// Return b, and leave the remaining unpolished in the buffer
			return len(b), nil
		}

	} else {
		// Read from the connection directly into the provided slice "b"
		return sconn.conn.Read(b)
	}
}

func (sconn *Connection) Write(b []byte) (int, error) {

	if sconn.state.polish != nil {
		// Polish data
		unpolished := b
		polished, polishError := sconn.state.polish.Polish(unpolished)
		if polishError != nil {
			return 0, polishError
		}

		numberOfBytesToWrite := len(polished)
		totalBytesWritten := 0

		// Write all of the bytes
		for numberOfBytesToWrite > totalBytesWritten {

			//Write the bytes in polish slice
			bytesWritten, writeError := sconn.conn.Write(polished)
			if writeError != nil {
				return bytesWritten, writeError
			}

			// Keep track of how many bytes we've written so far
			totalBytesWritten = totalBytesWritten + bytesWritten

			//If the bytes written are less than the bytes we need to write
			if numberOfBytesToWrite > totalBytesWritten {

				//Slice off what has already been written
				polished = polished[:bytesWritten]
			}
		}

		// Return the count of the pre-polished bytes
		return len(b), nil
	} else {
		return sconn.conn.Write(b)
	}
}

func (sconn *Connection) Close() error {
	return sconn.conn.Close()
}

func (sconn *Connection) LocalAddr() net.Addr {
	return sconn.conn.LocalAddr()
}

func (sconn *Connection) RemoteAddr() net.Addr {
	return sconn.conn.RemoteAddr()
}

func (sconn *Connection) SetDeadline(t time.Time) error {
	return sconn.conn.SetDeadline(t)
}

func (sconn *Connection) SetReadDeadline(t time.Time) error {
	return sconn.conn.SetReadDeadline(t)
}

func (sconn *Connection) SetWriteDeadline(t time.Time) error {
	return sconn.conn.SetWriteDeadline(t)
}

var _ net.Listener = (*replicantTransportListener)(nil)
var _ net.Conn = (*Connection)(nil)
