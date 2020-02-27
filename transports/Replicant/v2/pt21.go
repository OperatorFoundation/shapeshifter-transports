package replicant

import (
	pt "github.com/OperatorFoundation/shapeshifter-ipc"
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

	// FIXME - we need a real server config, not this empty one
	//config := ServerConfig{}
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
		polished := make([]byte, sconn.state.polish.GetChunkSize())

		// Read encrypted data from the connection and put it into our polished slice
		_, err := sconn.conn.Read(polished)
		if err != nil {
			return 0, err
		}

		// Decrypt the data
		unpolished, unpolishError := sconn.state.polish.Unpolish(polished)
		if unpolishError != nil {
			return 0, unpolishError
		}

		// Empty the buffer and write the decrypted data to it
		sconn.receiveBuffer.Reset()
		sconn.receiveBuffer.Write(unpolished)

		// Read the decrypted data into the provided slice "b"
		_, readError := sconn.receiveBuffer.Read(b)
		if readError != nil {
			return 0, readError
		}
		sconn.receiveBuffer.Reset()

		return len(b), nil
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
