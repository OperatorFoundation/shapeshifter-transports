package replicant

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/replicant/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/replicant/toneburst"
)

type ClientConfig struct {
	Toneburst toneburst.Config
	Polish    polish.ClientConfig
}

type ServerConfig struct {
	Toneburst toneburst.Config
	Polish    polish.ServerConfig
}
