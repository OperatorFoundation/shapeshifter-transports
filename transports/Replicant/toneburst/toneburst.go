package toneburst

import (
	"net"
)

type ToneBurst interface {
	Perform(conn net.Conn) error
}
