package replicant

import (
	"bytes"
	"fmt"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/polish"
	"golang.org/x/net/proxy"
	"testing"
)

func TestReplicantTransport_Dial(t *testing.T) {
	dialer := proxy.Direct
	replicantConfig := Config{
		Toneburst: nil,
		Polish:    nil,
	}
	replicantTransport := Transport{
		Config:  replicantConfig,
		Address: "127.0.0.1:1234",
		Dialer:  dialer,
	}

	_, err := replicantTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

// Polish Tests

// Silver
func TestNewSilverConfigs(t *testing.T) {
	silverServerConfig := polish.NewSilverServerConfig()
	if silverServerConfig == nil {
		t.Fail()
	}

	silverClientConfig  := polish.NewSilverClientConfig(silverServerConfig)
	if silverClientConfig == nil {
		t.Fail()
	}
}

func TestNewSilverClient(t *testing.T) {
	silverServerConfig := polish.NewSilverServerConfig()
	if silverServerConfig == nil {
		t.Fail()
	}

	silverClientConfig  := polish.NewSilverClientConfig(silverServerConfig)
	if silverClientConfig == nil {
		t.Fail()
	}

	silverClient := polish.NewSilverClient(*silverClientConfig)

	if silverClient == nil {
		t.Fail()
	}
}

func TestNewSilverServer(t *testing.T) {
	silverServerConfig := polish.NewSilverServerConfig()
	if silverServerConfig == nil {
		t.Fail()
	}

	silverServer := polish.NewSilverServer(*silverServerConfig)
	if silverServer == nil {
		t.Fail()
	}
}

func TestNewSilverServerConnection(t *testing.T) {
	silverServerConfig := polish.NewSilverServerConfig()
	if silverServerConfig == nil {
		t.Fail()
	}

	silverServer := polish.NewSilverServer(*silverServerConfig)
	if silverServer == nil {
		t.Fail()
	}
	// FIXME needs a connection
	//polishConnection := silverServer.NewConnection()
}

func TestSilverClientHandshake(t *testing.T) {

	silverServerConfig := polish.NewSilverServerConfig()
	if silverServerConfig == nil {
		t.Fail()
	}

	silverClientConfig  := polish.NewSilverClientConfig(silverServerConfig)
	if silverClientConfig == nil {
		t.Fail()
	}

	silverClient := polish.NewSilverClient(*silverClientConfig)

	if silverClient == nil {
		t.Fail()
	}

	//FIXME needs a connection
	//silverClient.Handshake()
}

func TestSilverPolishUnpolish(t *testing.T) {

	silverServerConfig := polish.NewSilverServerConfig()
	if silverServerConfig == nil {
		t.Fail()
	}

	silverClientConfig  := polish.NewSilverClientConfig(silverServerConfig)
	if silverClientConfig == nil {
		t.Fail()
	}

	silverClient := polish.NewSilverClient(*silverClientConfig)
	if silverClient == nil {
		t.Fail()
	}

	input := []byte{0, 1, 2, 3, 4}
	polished := silverClient.Polish(input)
	if bytes.Equal(input, polished) {
		fmt.Println("original input and polished are the same")
		t.Fail()
	}

	unpolished := silverClient.Unpolish(polished)
	if !bytes.Equal(unpolished, input) {
		fmt.Println("original input and unpolished are not the same")
		println(input)
		println(unpolished)
		t.Fail()
	}
}
