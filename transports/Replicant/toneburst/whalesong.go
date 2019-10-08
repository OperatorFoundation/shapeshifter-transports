package toneburst

import "net"

type WhalesongConfig struct {
	AddSequences    []Sequence
	RemoveSequences []Sequence
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
