package polish

import (
	"net"
)

type Connection interface {
	Handshake(conn net.Conn) error
	Polish(input []byte) []byte
	Unpolish(input []byte) []byte
}

type Server interface {
	NewConnection(net.Conn) Connection
}
