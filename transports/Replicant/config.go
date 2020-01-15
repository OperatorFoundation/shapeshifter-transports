package replicant

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/toneburst"
)

type ClientConfig struct {
	Toneburst toneburst.Config
	Polish    polish.ClientConfig
}

type ServerConfig struct {
	Toneburst toneburst.Config
	Polish    polish.ServerConfig
}
