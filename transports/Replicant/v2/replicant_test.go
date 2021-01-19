package replicant

import (
	"bytes"
	"github.com/OperatorFoundation/monolith-go/monolith"
	"github.com/OperatorFoundation/obfs4/common/log"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2/toneburst"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	runReplicantServer()

	os.Exit(m.Run())
}

func TestMarshalConfigs(t *testing.T) {
	clientConfig, serverConfig := createSilverMonotoneConfigsOneFixedAddByte()

	clientConfigJsonString, clientConfigJsonError := clientConfig.Marshal()
	if clientConfigJsonError != nil {
		t.Fail()
		return
	}

	serverConfigJsonString, serverConfigJsonError := serverConfig.Marshal()
	if serverConfigJsonError != nil {
		t.Fail()
		return
	}

	serverConfigWriteError := ioutil.WriteFile("ReplicantServerConfig.json", []byte(serverConfigJsonString), 0644)
	if serverConfigWriteError != nil {
		t.Fail()
		return
	}
	clientConfigWriteError := ioutil.WriteFile("ReplicantClientConfig.json", []byte(clientConfigJsonString), 0644)
	if clientConfigWriteError != nil {
		t.Fail()
		return
	}
}

func TestEmptyConfigs(t *testing.T) {
	clientConfig := ClientConfig{
		Toneburst: nil,
		Polish:    nil,
	}

	serverConfig := ServerConfig{
		Toneburst: nil,
		Polish:    nil,
	}

	replicantConnection(clientConfig, serverConfig, t)
}

func TestEmptyMonotone(t *testing.T) {
	clientConfig := createMonotoneClientConfigEmpty()
	serverConfig := createMonotoneServerConfigEmpty()
	replicantConnection(clientConfig, serverConfig, t)
}

func TestNilsMonotone(t *testing.T) {
	clientConfig := createMonotoneClientConfigNils()
	serverConfig := createMonotoneServerConfigNils()
	replicantConnection(clientConfig, serverConfig, t)
}

func MarshalSilverRandomEnumeratedConfigs(t *testing.T) {
	clientConfig, serverConfig := createSilverMonotoneConfigsRandomEnumeratedItems()

	clientConfigJsonString, clientConfigJsonError := clientConfig.Marshal()
	if clientConfigJsonError != nil {
		t.Fail()
		return
	}

	serverConfigJsonString, serverConfigJsonError := serverConfig.Marshal()
	if serverConfigJsonError != nil {
		t.Fail()
		return
	}

	serverConfigWriteError := ioutil.WriteFile("ReplicantServerConfig1.json", []byte(serverConfigJsonString), 0644)
	if serverConfigWriteError != nil {
		t.Fail()
		return
	}
	clientConfigWriteError := ioutil.WriteFile("ReplicantClientConfig1.json", []byte(clientConfigJsonString), 0644)
	if clientConfigWriteError != nil {
		t.Fail()
		return
	}
}

func TestOneFixedByteMonotone(t *testing.T) {
	clientConfig := createMonotoneClientConfigOneFixedAddByte()
	serverConfig := createMonotoneServerConfigOneFixedRemoveByte()
	replicantConnection(clientConfig, serverConfig, t)
}

func TestOneFixedByteSilverMonotone(t *testing.T) {
	clientConfig, serverConfig := createSilverMonotoneConfigsOneFixedAddByte()
	replicantConnection(*clientConfig, *serverConfig, t)
}

func TestOneAddOneRemove(t *testing.T) {
	clientConfig := createMonotoneClientConfigOneAddOneRemove()
	serverConfig := createMonotoneServerConfigOneAddOneRemove()
	replicantConnection(clientConfig, serverConfig, t)
}

func TestMonotoneEnumerated(t *testing.T) {
	clientConfig := createMonotoneClientConfigEnumeratedItems()
	serverConfig := createMonotoneServerConfigEnumeratedItems()
	replicantConnection(clientConfig, serverConfig, t)
}

func TestMonotoneRandom(t *testing.T) {
	clientConfig := createMonotoneClientConfigRandomItems()
	serverConfig := createMonotoneServerConfigRandomItems()
	replicantConnection(clientConfig, serverConfig, t)
}

func TestMonotoneRandomEnumerated(t *testing.T) {
	clientConfig := createMonotoneClientConfigRandomEnumeratedItems()
	serverConfig := createMonotoneServerConfigRandomEnumeratedItems()
	replicantConnection(clientConfig, serverConfig, t)
}

func TestSilver(t *testing.T) {
	clientConfig, serverConfig := createSilverConfigs()
	replicantConnection(*clientConfig, *serverConfig, t)
}

func runReplicantServer() {
	serverStarted := make(chan bool)
	addr := "127.0.0.1:1234"
	serverConfig := ServerConfig{
		Toneburst: nil,
		Polish:    nil,
	}

	go func() {
		listener := serverConfig.Listen(addr)
		serverStarted <- true

		lConn, lConnError := listener.Accept()
		if lConnError != nil {
			return
		}

		lBuffer := make([]byte, 4)
		_, lReadError := lConn.Read(lBuffer)
		if lReadError != nil {
			return
		}

		// Send a response back to person contacting us.
		_, lWriteError := lConn.Write([]byte("Message received."))
		if lWriteError != nil {
			return
		}
	}()

	serverFinishedStarting := <-serverStarted
	if !serverFinishedStarting {
	return
	}
}

// This test uses a more realistic config, like one might use in real deployment.
func TestSampleConfig(t *testing.T) {
	clientConfig, serverConfig := createSampleConfigs()
	replicantConnection(*clientConfig, *serverConfig, t)
}

func replicantConnection(clientConfig ClientConfig, serverConfig ServerConfig, t *testing.T) {
	serverStarted := make(chan bool)

	// Get a random port
	rand.Seed(time.Now().UnixNano())
	min := 1025
	max := 65535
	portNumber := min + rand.Intn(max-min+1)
	portString := strconv.Itoa(portNumber)
	addr := "127.0.0.1:"
	addr += portString

	go func() {
		listener := serverConfig.Listen(addr)
		serverStarted <- true

		lConn, lConnError := listener.Accept()
		if lConnError != nil {
			t.Fail()
			return
		}

		lBuffer := make([]byte, 4)
		_, lReadError := lConn.Read(lBuffer)
		if lReadError != nil {
			t.Fail()
			return
		}

		// Send a response back to person contacting us.
		_, lWriteError := lConn.Write([]byte("Message received."))
		if lWriteError != nil {
			t.Fail()
			return
		}

		_ = listener.Close()
	}()

	serverFinishedStarting := <-serverStarted
	if !serverFinishedStarting {
		t.Fail()
		return
	}

	cConn := clientConfig.Dial(addr)
	if cConn == nil {
		t.Fail()
		return
	}

	writeBytes := []byte{0x0A, 0x11, 0xB0, 0xB1}
	_, cWriteError := cConn.Write(writeBytes)
	if cWriteError != nil {
		t.Fail()
		return
	}

	readBuffer := make([]byte, 17)
	_, cReadError := cConn.Read(readBuffer)
	if cReadError != nil {
		t.Fail()
		return
	}

	_ = cConn.Close()

	return
}

func createMonotoneClientConfigNils() ClientConfig {
	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    nil,
		RemoveSequences: nil,
		SpeakFirst:      false,
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return clientConfig
}

func createMonotoneServerConfigNils() ServerConfig {
	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    nil,
		RemoveSequences: nil,
		SpeakFirst:      false,
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return serverConfig
}

func createMonotoneClientConfigEmpty() ClientConfig {
	parts := make([]monolith.Monolith, 0)
	desc := monolith.Description{Parts: parts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance
	removeSequences := desc

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeSequences,
		SpeakFirst:      false,
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return clientConfig
}

func createMonotoneServerConfigEmpty() ServerConfig {
	parts := make([]monolith.Monolith, 0)
	desc := monolith.Description{Parts: parts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance
	removeSequences := desc

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeSequences,
		SpeakFirst:      false,
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return serverConfig
}

func createMonotoneClientConfigOneFixedAddByte() ClientConfig {
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte: 0x13},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts: parts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: nil,
		SpeakFirst:      true,
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return clientConfig
}

func createMonotoneServerConfigOneFixedRemoveByte() ServerConfig {
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte: 0x13},
		},
	}
	parts = append(parts, part)

	desc := monolith.Description{Parts: parts}
	removeSequences := desc

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    nil,
		RemoveSequences: &removeSequences,
		SpeakFirst:      false,
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return serverConfig
}

func createMonotoneClientConfigOneAddOneRemove() ClientConfig {
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte: 0x13},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts: parts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance

	removeParts := make([]monolith.Monolith, 0)
	removePart := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte: 0x14},
		},
	}
	removeParts = append(removeParts, removePart)
	removeDesc := monolith.Description{Parts: removeParts}

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeDesc,
		SpeakFirst:      true,
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return clientConfig
}

func createMonotoneServerConfigOneAddOneRemove() ServerConfig {
	removeParts := make([]monolith.Monolith, 0)
	removePart := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte: 0x13},
		},
	}
	removeParts = append(removeParts, removePart)

	desc := monolith.Description{Parts: removeParts}
	removeSequences := desc

	addParts := make([]monolith.Monolith, 0)
	addPart := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte: 0x14},
		},
	}
	addParts = append(addParts, addPart)
	addDesc := monolith.Description{Parts: addParts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: addDesc,
		Args: args,
	}

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &monolithInstance,
		RemoveSequences: &removeSequences,
		SpeakFirst:      false,
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return serverConfig
}

func createMonotoneClientConfigEnumeratedItems() ClientConfig {
	set := []byte{0x11, 0x12, 0x13, 0x14}
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.EnumeratedByteType{Options: set},
			monolith.EnumeratedByteType{Options: set},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.EnumeratedByteType{Options: set},
			monolith.EnumeratedByteType{Options: set},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts: parts}
	bargs := []byte{0x14, 0x14, 0x14, 0x14}
	iargs := make([]interface{}, len(bargs))
	for index, value := range bargs {
		iargs[index] = value
	}
	args := monolith.NewArgs(iargs)
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance
	removeSequences := desc

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeSequences,
		SpeakFirst:      true,
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return clientConfig
}

func createMonotoneServerConfigEnumeratedItems() ServerConfig {
	rand.Seed(time.Now().UnixNano())
	set := []byte{0x11, 0x12, 0x13, 0x14}
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.EnumeratedByteType{Options: set},
			monolith.EnumeratedByteType{Options: set},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.EnumeratedByteType{Options: set},
			monolith.EnumeratedByteType{Options: set},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts: parts}
	bargs := []byte{0x11, 0x11, 0x11, 0x12}
	iargs := make([]interface{}, len(bargs))
	for index, value := range bargs {
		iargs[index] = value
	}
	args := monolith.NewArgs(iargs)
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance
	removeSequences := desc

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeSequences,
		SpeakFirst:      false,
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return serverConfig
}

func createMonotoneClientConfigRandomItems() ClientConfig {

	rand.Seed(time.Now().UnixNano())
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomByteType{},
			monolith.RandomByteType{},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomByteType{},
			monolith.RandomByteType{},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts: parts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance
	removeSequences := desc

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeSequences,
		SpeakFirst:      true,
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return clientConfig
}

func createMonotoneServerConfigRandomItems() ServerConfig {
	rand.Seed(time.Now().UnixNano())
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomByteType{},
			monolith.RandomByteType{},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomByteType{},
			monolith.RandomByteType{},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts: parts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance
	removeSequences := desc

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeSequences,
		SpeakFirst:      false,
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return serverConfig
}

func createSilverMonotoneConfigsRandomEnumeratedItems() (ClientConfig, ServerConfig) {
	serverConfig := createMonotoneServerConfigRandomEnumeratedItems()
	clientConfig := createMonotoneClientConfigRandomEnumeratedItems()

	silverServerConfig, _ := polish.NewSilverServerConfig()
	silverClientConfig, _ := polish.NewSilverClientConfig(silverServerConfig)

	serverConfig.Polish = silverServerConfig
	clientConfig.Polish = silverClientConfig

	return clientConfig, serverConfig
}

func createMonotoneClientConfigRandomEnumeratedItems() ClientConfig {
	rand.Seed(time.Now().UnixNano())
	set := []byte{0x11, 0x12, 0x13, 0x14}
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{RandomOptions: set},
			monolith.RandomEnumeratedByteType{RandomOptions: set},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{RandomOptions: set},
			monolith.RandomEnumeratedByteType{RandomOptions: set},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts: parts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance
	removeSequences := desc

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeSequences,
		SpeakFirst:      true,
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return clientConfig
}

func createMonotoneServerConfigRandomEnumeratedItems() ServerConfig {
	rand.Seed(time.Now().UnixNano())
	set := []byte{0x11, 0x12, 0x13, 0x14}
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{RandomOptions: set},
			monolith.RandomEnumeratedByteType{RandomOptions: set},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{RandomOptions: set},
			monolith.RandomEnumeratedByteType{RandomOptions: set},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts: parts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance
	removeSequences := desc

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeSequences,
		SpeakFirst:      false,
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneConfig,
		Polish:    nil,
	}

	return serverConfig
}

func createSilverConfigs() (*ClientConfig, *ServerConfig) {
	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		return nil, nil
	}
	polishClientConfig, polishClientError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientError != nil {
		return nil, nil
	}

	clientConfig := ClientConfig{
		Toneburst: nil,
		Polish:    polishClientConfig,
	}

	serverConfig := ServerConfig{
		Toneburst: nil,
		Polish:    polishServerConfig,
	}

	return &clientConfig, &serverConfig
}

func createSilverMonotoneConfigsOneFixedAddByte() (*ClientConfig, *ServerConfig) {
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte: 0x13},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts: parts}
	args := monolith.NewEmptyArgs()
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &monolithInstance,
		RemoveSequences: nil,
		SpeakFirst:      true,
	}

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		return nil, nil
	}

	polishClientConfig, polishClientConfigError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientConfigError != nil {
		return nil, nil
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneConfig,
		Polish:    polishClientConfig,
	}

	serverParts := make([]monolith.Monolith, 0)
	serverPart := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte: 0x13},
		},
	}
	serverParts = append(serverParts, serverPart)

	serverDesc := monolith.Description{Parts: serverParts}
	serverRemoveSequences := serverDesc

	monotoneServerConfig := toneburst.MonotoneConfig{
		AddSequences:    nil,
		RemoveSequences: &serverRemoveSequences,
		SpeakFirst:      false,
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneServerConfig,
		Polish:    polishServerConfig,
	}

	return &clientConfig, &serverConfig
}

// Polish Tests
func TestPolishOnlyConnection(t *testing.T) {
	clientConfig, serverConfig := createSilverConfigs()
	replicantConnection(*clientConfig, *serverConfig, t)
}

//Both
func TestWithSilverMonotone(t *testing.T) {
	clientConfig, serverConfig := createSilverMonotoneConfigsRandomEnumeratedItems()
	replicantConnection(clientConfig, serverConfig, t)
}

func TestSilverClientPolishUnpolish(t *testing.T) {
	silverServerConfig, serverConfigError := polish.NewSilverServerConfig()
	if serverConfigError != nil {
		t.Fail()
		return
	}

	if silverServerConfig == nil {
		t.Fail()
		return
	}

	silverClientConfig, clientConfigError := polish.NewSilverClientConfig(silverServerConfig)
	if clientConfigError != nil {
		t.Fail()
		return
	}

	if silverClientConfig == nil {
		t.Fail()
		return
	}

	silverClient, clientError := polish.NewSilverClient(*silverClientConfig)
	if clientError != nil {
		t.Fail()
		return
	}

	if silverClient == nil {
		t.Fail()
		return
	}

	input := []byte{3, 12, 2, 6, 31}

	polished, polishError := silverClient.Polish(input)
	if polishError != nil {
		t.Fail()
		return
	}

	if bytes.Equal(input, polished) {
		t.Fail()
		return
	}

	unpolished, unpolishError := silverClient.Unpolish(polished)
	if unpolishError != nil {
		t.Fail()
		return
	}

	if !bytes.Equal(unpolished, input) {
		t.Fail()
		return
	}
}

func createSampleConfigs() (*ClientConfig, *ServerConfig) {
	rand.Seed(time.Now().UnixNano())

	clientParts := []monolith.Monolith{
		monolith.BytesPart{
			Items: []monolith.ByteType{
				monolith.SemanticIntProducerByteType{Name: "m", Value: monolith.RandomByteType{}},
			},
		},
		&monolith.SemanticSeedConsumerDynamicPart{Name: "m", Item: monolith.RandomByteType{}},
	}

	clientDesc := monolith.Description{Parts: clientParts}

	clientInstance := monolith.Instance{
		Desc: clientDesc,
		Args: monolith.NewEmptyArgs(),
	}

	serverParts := []monolith.Monolith{
		monolith.BytesPart{
			Items: []monolith.ByteType{
				monolith.SemanticIntProducerByteType{Name: "n", Value: monolith.RandomByteType{}},
			},
		},
		&monolith.SemanticSeedConsumerDynamicPart{Name: "n", Item: monolith.RandomByteType{}},
	}

	serverDesc := monolith.Description{Parts: serverParts}

	serverInstance := monolith.Instance{
		Desc: serverDesc,
		Args: monolith.NewEmptyArgs(),
	}

	monotoneServerConfig := toneburst.MonotoneConfig{
		AddSequences:    &serverInstance,
		RemoveSequences: &clientDesc,
		SpeakFirst:      false,
	}

	monotoneClientConfig := toneburst.MonotoneConfig{
		AddSequences:    &clientInstance,
		RemoveSequences: &serverDesc,
		SpeakFirst:      true,
	}

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		return nil, nil
	}

	polishClientConfig, polishClientConfigError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientConfigError != nil {
		return nil, nil
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneClientConfig,
		Polish:    polishClientConfig,
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneServerConfig,
		Polish:    polishServerConfig,
	}

	return &clientConfig, &serverConfig
}

func TestRoundTrip(test *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	listener, listenError := net.Listen("tcp", "localhost:0")
	require.NoError(test, listenError)

	encodedServerConfig, serverConfigError := ioutil.ReadFile("RoundTripServerConfig.txt")
	require.NoError(test, serverConfigError)
	encodedClientConfig, clientConfigError := ioutil.ReadFile("RoundTripClientConfig.txt")
	require.NoError(test, clientConfigError)

	listener, listenError = Wrap(listener, string(encodedServerConfig))
	require.NoError(test, listenError)
	defer listener.Close()
	log.Debugf("Listening at %v", listener.Addr())

	go func() {
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}

			for {
				byteArray := make([]byte, 2)
				bytesRead, readError := conn.Read(byteArray)
				if readError != nil {
					break
				}
				if bytesRead <= 0 {
					break
				}
				bytesWritten, writeError := conn.Write(byteArray)
				if writeError != nil {
					break
				}
				if bytesWritten <= 0 {
					break
				}
			}
		}
	}()

	config, decodeError := DecodeClientConfig(string(encodedClientConfig))
	require.NoError(test, decodeError)
	dial := func(addr string) (net.Conn, error) {
		clientConnection, dialError := net.Dial("tcp", addr)
		if dialError != nil {
			return nil, dialError
		}
		return NewClientConnection(clientConnection, *config)
	}

	iters := 1
	var waitGroup sync.WaitGroup
	waitGroup.Add(iters)
	for i := 0; i < iters; i++ {
		go func() {

			defer waitGroup.Done()
			dialConn, dialErr := dial(listener.Addr().String())
			require.NoError(test, dialErr)
			defer dialConn.Close()
			checkBytes(dialConn, test)
			checkBytes(dialConn, test)
			checkBytes(dialConn, test)
		}()
	}

	waitGroup.Wait()
}

func checkBytes(dialConn net.Conn, test *testing.T) {
	byteBuf1 := make([]byte, 2)
	rand.Read(byteBuf1)
	bytesWritten, writeErr := dialConn.Write(byteBuf1)
	require.NoError(test, writeErr)
	assert.EqualValues(test, 2, bytesWritten)

	byteBuf2 := make([]byte, 2)
	bytesRead, readErr := io.ReadFull(dialConn, byteBuf2)
	require.NoError(test, readErr)
	assert.EqualValues(test, 2, bytesRead)
	assert.EqualValues(test, string(byteBuf1), string(byteBuf2))
}

func Wrap(ll net.Listener, encodedConfig string) (net.Listener, error) {
	serverConfig, err := DecodeServerConfig(encodedConfig)
	if err != nil {
		return nil, err
	}

	return &ReplicantListener{
		Listener: ll,
		cfg:      serverConfig,
	}, nil
}

type ReplicantListener struct {
	net.Listener

	cfg *ServerConfig
}

func (l ReplicantListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return NewServerConnection(conn, *l.cfg)
}