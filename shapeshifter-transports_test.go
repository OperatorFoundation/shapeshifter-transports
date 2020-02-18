package shapeshifter_transports

import (
	"fmt"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Dust"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Optimizer"
	replicant "github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/meeklite"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/meekserver"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow"
	"testing"
)


func TestTransports(t *testing.T) {
	shadowConfig := shadow.Config{}
	obfs4Config := obfs4.Config{}
	obfs2Config:= obfs2.Transport{}
	meekliteConfig := meeklite.Config{}
	meekserverConfig := meekserver.State{}
	ReplicantConfig := replicant.ClientConfig{}
	DustConfig := Dust.Config{}
	OptimizerConfig := Optimizer.Client{}
	fmt.Println(shadowConfig, obfs2Config, obfs4Config, meekliteConfig, meekserverConfig, ReplicantConfig, DustConfig, OptimizerConfig)
}
