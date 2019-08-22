package replicant

import (
	"fmt"
	pt "github.com/OperatorFoundation/shapeshifter-ipc"
	"net"
	"time"
)

// Create outgoing transport connection
func (transport *replicantTransport) Dial(address string) net.Conn {
	conn, dialErr := net.Dial("tcp", address)
	if dialErr != nil {
		return nil
	}

	config := Config{}

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
func (transport *replicantTransport) Listen(address string) net.Listener {
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

	return newReplicantTransportListener(ln, transport)
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

	config := Config{}

	return NewServerConnection(conn, config)
}

// Close closes the transport listener.
// Any blocked Accept operations will be unblocked and return errors.
func (listener *replicantTransportListener) Close() error {
	return listener.listener.Close()
}

func (sconn *ReplicantConnection) Read(b []byte) (int, error) {
	polished := b

	_, err := sconn.conn.Read(polished)
	if err != nil {
		return 0, err
	}

	unpolished := sconn.state.polish.Unpolish(polished)
	sconn.receiveBuffer.Reset()
	sconn.receiveBuffer.Write(unpolished)
	sconn.receiveBuffer.Read(b)
	sconn.receiveBuffer.Reset()

	// FIXME
	return len(b), nil
}

func (sconn *ReplicantConnection ) Write(b []byte) (int, error) {
	unpolished := b
	polished := sconn.state.polish.Polish(unpolished)

	return sconn.conn.Write(polished)
}

func (sconn *ReplicantConnection ) Close() error {
	return sconn.conn.Close()
}

func (sconn *ReplicantConnection ) LocalAddr() net.Addr {
	return sconn.conn.LocalAddr()
}

func (sconn *ReplicantConnection ) RemoteAddr() net.Addr {
	return sconn.conn.RemoteAddr()
}

func (sconn *ReplicantConnection ) SetDeadline(t time.Time) error {
	return sconn.conn.SetDeadline(t)
}

func (sconn *ReplicantConnection ) SetReadDeadline(t time.Time) error {
	return sconn.conn.SetReadDeadline(t)
}

func (sconn *ReplicantConnection ) SetWriteDeadline(t time.Time) error {
	return sconn.conn.SetWriteDeadline(t)
}

var _ net.Listener = (*replicantTransportListener)(nil)
var _ net.Conn = (*ReplicantConnection )(nil)
