package toneburst

import (
	"fmt"
	"net"
)

type WhalesongConfig struct {
	AddSequences []Sequence
	RemoveSequences []Sequence
}

func (config WhalesongConfig) Construct() (ToneBurst, error) {
	return NewWhalesong(config), nil
}

type Sequence []byte

type Whalesong struct {
	config WhalesongConfig
}

func NewWhalesong(config WhalesongConfig) *Whalesong {
	return &Whalesong{config: config}
}

func (whalesong *Whalesong) Perform(conn net.Conn) error {
	fmt.Println(conn)
	return nil
}
