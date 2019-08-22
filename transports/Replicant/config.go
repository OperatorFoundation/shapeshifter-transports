package replicant

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/replicant/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/replicant/toneburst"
)

type Config struct {
	toneburst toneburst.Config
	polish    polish.Config
}
