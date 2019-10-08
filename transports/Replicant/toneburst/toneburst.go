package toneburst

import (
	"net"
)

type ToneBurst interface {
	Perform(conn net.Conn) error
}

func New(config *Config) ToneBurst {
	if config == nil {
		return nil
	}
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
