package toneburst

import "net"

type WhalesongConfig struct {
	addSequences []Sequence
	removeSequences []Sequence
}

type Sequence []byte

type Whalesong struct {
	config WhalesongConfig
}

func NewWhalesong(config WhalesongConfig) *Whalesong {
	return &Whalesong{config: config}
}

func (whalesong *Whalesong) Perform(conn net.Conn) error {
	return nil
}
