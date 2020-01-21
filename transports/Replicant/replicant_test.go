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
	replicantConfig := ClientConfig{
		Toneburst: nil,
		Polish:    nil,
	}

	replicantTransport := Transport{
		Config:  replicantConfig,
		Address: "159.203.158.90:1234",
		Dialer:  dialer,
	}

	_, err := replicantTransport.Dial()
	if err != nil {
		println(err.Error())
		t.Fail()
	}
}

// Polish Tests

// Silver
func TestNewSilverConfigs(t *testing.T) {
	silverServerConfig, serverConfigError := polish.NewSilverServerConfig()

	if serverConfigError != nil {
		println("Silver server config error: ", serverConfigError.Error())
		t.Fail()
	}
	if silverServerConfig == nil {
		t.Fail()
	}

	silverClientConfig, clientConfigError  := polish.NewSilverClientConfig(silverServerConfig)

	if clientConfigError != nil {
		println("Silver client config error: ", clientConfigError)
		t.Fail()
	}

	if silverClientConfig == nil {
		t.Fail()
	}
}

func TestNewSilverClient(t *testing.T) {
	silverServerConfig, serverConfigError := polish.NewSilverServerConfig()
	if serverConfigError != nil{
		println("Silver server config error: ", serverConfigError.Error())
		t.Fail()
	}

	if silverServerConfig == nil {
		t.Fail()
	}

	silverClientConfig, clientConfigError  := polish.NewSilverClientConfig(silverServerConfig)

	if clientConfigError != nil {
		println("Silver client config error: ", clientConfigError)
		t.Fail()
	}

	if silverClientConfig == nil {
		t.Fail()
	}

	silverClient, clientError := polish.NewSilverClient(*silverClientConfig)

	if clientError != nil {
		println("Silver client error: ", clientError)
		t.Fail()
	}

	if silverClient == nil {
		t.Fail()
	}
}

func TestNewSilverServer(t *testing.T) {
	silverServerConfig, serverConfigError := polish.NewSilverServerConfig()

	if serverConfigError != nil{
		println("Silver server config error: ", serverConfigError.Error())
		t.Fail()
	}

	if silverServerConfig == nil {
		t.Fail()
	}

	_, serverError := polish.NewSilverServer(*silverServerConfig)

	if serverError != nil {
		println("Silver server error: ", serverError)
		t.Fail()
	}
}

func TestNewSilverServerConnection(t *testing.T) {
	silverServerConfig, serverConfigError := polish.NewSilverServerConfig()
	if silverServerConfig == nil {
		t.Fail()
	}

	if serverConfigError != nil{
		println("Silver server config error: ", serverConfigError.Error())
		t.Fail()
	}

	_, serverError := polish.NewSilverServer(*silverServerConfig)

	if serverError != nil {
		println("Silver server error: ", serverError)
		t.Fail()
	}
	// FIXME needs a connection
	//polishConnection := silverServer.NewConnection()
}

func TestSilverClientHandshake(t *testing.T) {

	silverServerConfig, serverConfigError := polish.NewSilverServerConfig()
	if silverServerConfig == nil {
		t.Fail()
	}

	if serverConfigError != nil{
		println("Silver server config error: ", serverConfigError.Error())
		t.Fail()
	}

	silverClientConfig, clientConfigError  := polish.NewSilverClientConfig(silverServerConfig)
	if silverClientConfig == nil {
		t.Fail()
	}

	if clientConfigError != nil {
		println("Silver client config error: ", clientConfigError)
		t.Fail()
	}

	silverClient, clientError := polish.NewSilverClient(*silverClientConfig)

	if clientError != nil {
		println("Silver client error: ", clientError)
		t.Fail()
	}

	if silverClient == nil {
		t.Fail()
	}

	//FIXME needs a connection
	//silverClient.Handshake()
}

func TestSilverPolishUnpolish(t *testing.T) {

	silverServerConfig, serverConfigError := polish.NewSilverServerConfig()

	if serverConfigError != nil{
		println("Silver server config error: ", serverConfigError.Error())
		t.Fail()
	}

	if silverServerConfig == nil {
		t.Fail()
	}

	silverClientConfig, clientConfigError  := polish.NewSilverClientConfig(silverServerConfig)
	if silverClientConfig == nil {
		t.Fail()
	}

	if clientConfigError != nil {
		println("Silver client config error: ", clientConfigError)
		t.Fail()
	}

	silverClient, clientError := polish.NewSilverClient(*silverClientConfig)

	if clientError != nil {
		println("Silver client error: ", clientError)
		t.Fail()
	}

	if silverClient == nil {
		t.Fail()
	}

	input := []byte{0, 1, 2, 3, 4}

	polished, polishError := silverClient.Polish(input)
	if polishError != nil {
		println("Received polish error: ", polishError)
		t.Fail()
	}

	if bytes.Equal(input, polished) {
		fmt.Println("original input and polished are the same")
		t.Fail()
	}

	println("data before polish length:", len(input))
	println("after polish: ", len(polished))

	unpolished, unpolishError := silverClient.Unpolish(polished)
	if unpolishError != nil {
		println("Received an unpolish error: ", unpolishError)
		t.Fail()
	}

	println("unpolished length: ", len(unpolished))
	if !bytes.Equal(unpolished, input) {
		fmt.Println("original input and unpolished are not the same")
		fmt.Printf("%v\n", input)
		fmt.Printf("%v\n", unpolished)
		t.Fail()
	}
}
