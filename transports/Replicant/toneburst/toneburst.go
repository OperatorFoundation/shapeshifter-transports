package toneburst

import (
	"net"
)

type ToneBurst interface {
	Perform(conn net.Conn) error
}

func New(config Config) interface{ToneBurst} {
	switch config.Selector {
	case "whalesong":
		if config.Whalesong == nil {
			return nil
		} else {
			return NewWhalesong(*config.Whalesong)
		}
	default:
		return nil
	}
}
