package toneburst

import "net"

type MonolithConfig struct {
	AddSequences    []Sequence
	RemoveSequences []Sequence
}

type Monolith struct {
	config MonolithConfig
}

func NewMonolith(config MonolithConfig) *Monolith {
	return &Monolith{config: config}
}

func (monolith *Monolith) Perform(conn net.Conn) error {
	return nil
}