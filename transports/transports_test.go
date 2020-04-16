package transports

import (
	"fmt"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Dust/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Optimizer/v2"
	replicant "github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/meeklite/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/meekserver/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs2/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow/v2"

	"testing"
)


func TestTransports(t *testing.T) {
	shadowConfig := shadow.Config{}
	obfs4Config := obfs4.Config{}
	obfs2Config:= obfs2.Transport{}
	meekliteConfig := meeklite.Config{}
	meekserverConfig := meekserver.Config{}
	ReplicantConfig := replicant.ClientConfig{}
	DustConfig := Dust.Config{}
	OptimizerConfig := Optimizer.Client{}
	fmt.Println(shadowConfig, obfs2Config, obfs4Config, meekliteConfig, meekserverConfig, ReplicantConfig, DustConfig, OptimizerConfig)
}
