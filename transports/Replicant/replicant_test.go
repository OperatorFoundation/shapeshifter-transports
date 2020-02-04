package replicant

import (
	"bytes"
	"fmt"
	"github.com/OperatorFoundation/monolith-go/monolith"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/toneburst"
	"math/rand"
	"testing"
	"time"
)

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

func TestOneFixedByteMonotone(t *testing.T) {
	clientConfig := createMonotoneClientConfigOneFixedAddByte()
	serverConfig := createMonotoneServerConfigOneFixedRemoveByte()
	replicantConnection(clientConfig, serverConfig, t)
}

func TestOneAddOneRemove(t *testing.T) {
	clientConfig := createMonotoneClientConfigOneAddOneRemove()
	serverConfig := createMonotoneServerConfigOneAddOneRemove()
	replicantConnection(clientConfig, serverConfig, t)
}

func TestMonotoneRandomEnumerated(t *testing.T) {
	clientConfig := createMonotoneClientConfigRandomEnumeratedItems()
	serverConfig := createMonotoneServerConfigRandomEnumeratedItems()
	replicantConnection(clientConfig, serverConfig, t)
}

func replicantConnection(clientConfig ClientConfig, serverConfig ServerConfig, t *testing.T) {
	serverStarted := make(chan bool)

	go func() {
		listener := serverConfig.Listen("127.0.0.1:7777")
		defer listener.Close()
		println("Created listener")
		serverStarted <- true

		for {
			lConn, lConnError := listener.Accept()
			if lConnError != nil {
				println("Listener connection error: ", lConnError.Error())
				t.Fail()
				return
			}

			println("received an incoming connection")
			lBuffer := make([]byte, 1024)
			lReadLength, lReadError := lConn.Read(lBuffer)
			if lReadError != nil {
				println("Listener read error: ", lReadError)
				t.Fail()
				return
			}

			println("Listener read length: ", lReadLength)
			// Send a response back to person contacting us.
			lWriteLength, lWriteError := lConn.Write([]byte("Message received."))
			if lWriteError != nil {
				println("Listener write error: ", lWriteError)
				t.Fail()
				return
			}
			println("Wrote a response to the client. Length: ", lWriteLength)
		}
	}()

	serverFinishedStarting := <- serverStarted
	if !serverFinishedStarting {
		t.Fail()
		return
	}

	cConn := clientConfig.Dial("127.0.0.1:7777")
	if cConn == nil {
		println("Dial error: client connection is nil.")
		t.Fail()
		return
	}
	println("Created client connection.")

	writeBytes := []byte{0x0A, 0x11, 0xB0, 0xB1}
	cWriteLength, cWriteError := cConn.Write(writeBytes)
	if cWriteError != nil {
		println("Client write error: ", cWriteError)
		t.Fail()
		return
	}
	println("Wrote bytes to the server, count: ", cWriteLength)

	readBuffer := make([]byte, 1024)
	cReadLength, cReadError := cConn.Read(readBuffer)
	if cReadError != nil {
		println("Client read error: ", cReadError)
		t.Fail()
		return
	}
	println("Client read byte count: ", cReadLength)
	fmt.Printf("Client read buffer: %v:\n", readBuffer)

	defer cConn.Close()

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
	desc := monolith.Description{parts}
	args := make([]interface{}, 0)
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
	desc := monolith.Description{parts}
	args := make([]interface{}, 0)
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
			monolith.FixedByteType{Byte:0x13},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts:parts}
	args := make([]interface{}, 0)
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
			monolith.FixedByteType{Byte:0x13},
		},
	}
	parts = append(parts, part)

	desc := monolith.Description{Parts:parts}
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
			monolith.FixedByteType{Byte:0x13},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{Parts:parts}
	args := make([]interface{}, 0)
	monolithInstance := monolith.Instance{
		Desc: desc,
		Args: args,
	}

	addSequences := monolithInstance

	removeParts := make([]monolith.Monolith, 0)
	removePart := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte:0x14},
		},
	}
	removeParts = append(removeParts, removePart)
	removeDesc := monolith.Description{Parts:removeParts}

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
			monolith.FixedByteType{Byte:0x13},
		},
	}
	removeParts = append(removeParts, removePart)

	desc := monolith.Description{Parts: removeParts}
	removeSequences := desc

	addParts := make([]monolith.Monolith, 0)
	addPart := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.FixedByteType{Byte:0x14},
		},
	}
	addParts = append(addParts, addPart)
	addDesc := monolith.Description{Parts:addParts}
	args := make([]interface{}, 0)
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

func createMonotoneClientConfigRandomEnumeratedItems() ClientConfig {
	rand.Seed(time.Now().UnixNano())
	set := []byte{0x11, 0x12, 0x13, 0x14}
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{set},
			monolith.RandomEnumeratedByteType{set},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{set},
			monolith.RandomEnumeratedByteType{set},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{parts}
	args := make([]interface{}, 0)
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
			monolith.RandomEnumeratedByteType{set},
			monolith.RandomEnumeratedByteType{set},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{set},
			monolith.RandomEnumeratedByteType{set},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{parts}
	args := make([]interface{}, 0)
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

func createSilverConfigs()(*ClientConfig, *ServerConfig) {
	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		println("Polish server error: ", polishServerError)
		return nil, nil
	}
	polishClientConfig, polishClientError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientError != nil {
		println("Polish  client error: ", polishClientError)
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

func createSilverMonotoneClientConfigRandomEnumeratedItems() *ClientConfig {
	rand.Seed(time.Now().UnixNano())
	set := []byte{0x11, 0x12, 0x13, 0x14}
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{set},
			monolith.RandomEnumeratedByteType{set},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{set},
			monolith.RandomEnumeratedByteType{set},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{parts}
	args := make([]interface{}, 0)
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

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		println("Polish server error: ", polishServerError)
		return nil
	}

	polishClientConfig, polishClientConfigError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientConfigError != nil {
		println("Error creating silver client config: ", polishClientConfigError)
		return nil
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneConfig,
		Polish:    polishClientConfig,
	}

	return &clientConfig
}

func createSilverMonotoneServerConfigRandomEnumeratedItems() *ServerConfig {
	rand.Seed(time.Now().UnixNano())
	set := []byte{0x11, 0x12, 0x13, 0x14}
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{set},
			monolith.RandomEnumeratedByteType{set},
		},
	}
	parts = append(parts, part)
	part = monolith.BytesPart{
		Items: []monolith.ByteType{
			monolith.RandomEnumeratedByteType{set},
			monolith.RandomEnumeratedByteType{set},
		},
	}
	parts = append(parts, part)
	desc := monolith.Description{parts}
	args := make([]interface{}, 0)
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

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		println("Polish server error: ", polishServerError)
		return nil
	}

	serverConfig := ServerConfig{
		Toneburst: monotoneConfig,
		Polish:    polishServerConfig,
	}

	return &serverConfig
}

// Polish Tests
func TestConn(t *testing.T) {
	clientConfig, serverConfig := createSilverConfigs()
	// Run the server concurrently
	go func() {
		listener := serverConfig.Listen("127.0.0.1:7777")
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				println("Listener accept func error: ", err)
				return
			}
			defer conn.Close()

			println("received an incoming connection")
			// Make a buffer to hold incoming data.
			buf := make([]byte, 1024)
			// Read the incoming connection into the buffer.
			reqLen, err := conn.Read(buf)
			if err != nil {
				fmt.Println("Error reading:", err.Error())
			}

			println("regLen: ", reqLen)
			println("readBuffer: ")
			fmt.Println(string(buf[:]))
			// Send a response back to person contacting us.
			conn.Write([]byte("Message received."))
		}
	}()

	// Run the client
	client := clientConfig.Dial("127.0.0.1:7777")
	if client == nil {
		println("Dial error: Conn is nil.")
		t.Fail()
		return
	}

	println("Successful client connection to a replicant server with Silver polish!")
	writeBytes := []byte{0x0A, 0x11, 0xB0, 0xB1}
	writeCount, writeError := client.Write(writeBytes)
	if writeError != nil {
		println("Write error from client: ", writeError)
	}
	println("Client write byte count = ", writeCount)

	readBuffer := make([]byte, 1024)
	client.Read(readBuffer)
	fmt.Printf("Client read buffer: %v:\n", readBuffer)

	defer client.Close()
	return
}

func TestPolishOnlyConnection(t *testing.T) {
	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		t.Fail()
		return
	}

	go runServerWithSilver(polishServerConfig)
	runClientWithSilver(polishServerConfig)
}

func TestPolishOnlyClientSend(t *testing.T) {
	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		t.Fail()
		return
	}

	go runServerWithSilver(polishServerConfig)
	clientWriteSilver(polishServerConfig)
}

func runServerWithSilver(polishServerConfig *polish.SilverPolishServerConfig) {

	serverConfig := ServerConfig{
		Toneburst: nil,
		Polish:    polishServerConfig,
	}

	listener := serverConfig.Listen("127.0.0.1:7777")
	_, serverConnError := listener.Accept()
	if serverConnError != nil {
		return
	}
}

//Both
func TestWithSilverMonotone(t *testing.T) {

	clientConfig := createSilverMonotoneClientConfigRandomEnumeratedItems()
	serverConfig := createSilverMonotoneServerConfigRandomEnumeratedItems()
	//monotoneConfig := createMonotoneConfig()

	go func() {
		listener := serverConfig.Listen("127.0.0.1:7777")
		println("Created listener")
		defer listener.Close()

		for {
			lConn, lConnError := listener.Accept()
			if lConnError != nil {
				println("Listener connection error: ", lConnError.Error())
				t.Fail()
				return
			}

			println("received an incoming connection")
			lBuffer := make([]byte, 1024)
			lReadLength, lReadError := lConn.Read(lBuffer)
			if lReadError != nil {
				println("Listener read error: ", lReadError)
				t.Fail()
				return
			}
			println("Listener read length: ", lReadLength)
			// Send a response back to person contacting us.
			lWriteLength, lWriteError := lConn.Write([]byte("Message received."))
			if lWriteError != nil {
				println("Listener write error: ", lWriteError)
				t.Fail()
				return
			}
			println("Wrote a response to the client. Length: ", lWriteLength)
		}
	}()

	cConn := clientConfig.Dial("127.0.0.1:7777")
	if cConn == nil {
		println("Dial error: client connection is nil.")
		t.Fail()
		return
	}
	println("Created client connection.")

	writeBytes := []byte{0x0A, 0x11, 0xB0, 0xB1}
	cWriteLength, cWriteError := cConn.Write(writeBytes)
	if cWriteError != nil {
		println("Client write error: ", cWriteError)
		t.Fail()
		return
	}
	println("Wrote bytes to the server, count: ", cWriteLength)

	readBuffer := make([]byte, 1024)
	cReadLength, cReadError := cConn.Read(readBuffer)
	if cReadError != nil {
		println("Client read error: ", cReadError)
		t.Fail()
		return
	}
	println("Client read byte count: ", cReadLength)
	fmt.Printf("Client read buffer: %v:\n", readBuffer)

	defer cConn.Close()

	return
}

func runClientWithSilver(polishServerConfig *polish.SilverPolishServerConfig) {
	polishClientConfig, polishClientError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientError != nil {
		return
	}

	clientConfig := ClientConfig{
		Toneburst: nil,
		Polish:    polishClientConfig,
	}

	client := clientConfig.Dial("127.0.0.1:7777")
	if client == nil {
		return
	} else {
		println("Successful client connection to a replicant server with Silver polish!")
		readBuffer := make([]byte, 1024)
		client.Read(readBuffer)
	}
}

func clientWriteSilver(polishServerConfig *polish.SilverPolishServerConfig) {
	polishClientConfig, polishClientError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientError != nil {
		return
	}

	clientConfig := ClientConfig{
		Toneburst: nil,
		Polish:    polishClientConfig,
	}

	client := clientConfig.Dial("127.0.0.1:7777")
	if client == nil {
		return
	} else {
		println("Successful client connection to a replicant server with Silver polish!")
		readBuffer := make([]byte, 1024)
		go client.Read(readBuffer)
		writeBytes := []byte{0x0A, 0x11, 0xB0, 0xB1}
		writeCount, writeError := client.Write(writeBytes)
		if writeError != nil {
			println("Write error from client: ", writeError)
		}

		println("Write byte count = ", writeCount)
	}
}

func TestSilverClientPolishUnpolish(t *testing.T) {

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