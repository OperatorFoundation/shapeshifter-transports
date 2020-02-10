package replicant

import (
	"fmt"
	pt "github.com/OperatorFoundation/shapeshifter-ipc"
	"net"
	"time"
)

// Create outgoing transport connection
func (config ClientConfig) Dial(address string) net.Conn {
	conn, dialErr := net.Dial("tcp", address)
	if dialErr != nil {
		fmt.Println("Dial Error: ", dialErr)
		return nil
	}

	transportConn, err := NewClientConnection(conn, config)
	if err != nil {
		fmt.Println("Connection Error: ", err)
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
		fmt.Println(resolveErr.Error())
		return nil
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
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
		polished := b

		// Read encrypted data from the connection and put it into our polished slice
		_, err := sconn.conn.Read(polished)
		if err != nil {
			return 0, err
		}

		// Decrypt the data
		unpolished, unpolishError := sconn.state.polish.Unpolish(polished)
		if unpolishError != nil {
			println("Received an unpolish error: ", unpolishError.Error())
			println("Polished input: ", polished)
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
		unpolished := b
		polished, polishError := sconn.state.polish.Polish(unpolished)
		if polishError != nil {
			return 0, polishError
		}

		println("Polished data to write: ", polished)
		return sconn.conn.Write(polished)
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
