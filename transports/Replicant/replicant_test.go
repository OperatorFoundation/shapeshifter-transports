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

func TestMonotoneOnly(t *testing.T) {
	monotoneClientConfig := createMonotoneClientConfigRandomEnumeratedItems()
	monotoneServerConfig := createMonotoneServerConfigRandomEnumeratedItems()
	//monotoneConfig := createMonotoneConfig()

	serverConfig := ServerConfig{
		Toneburst: monotoneServerConfig,
		Polish:    nil,
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneClientConfig,
		Polish:    nil,
	}

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

func createMonotoneConfig() toneburst.MonotoneConfig {
	parts := make([]monolith.Monolith, 0)
	part := monolith.BytesPart{Items:[]monolith.ByteType{monolith.FixedByteType{Byte:0x0A}}}
	parts = append(parts, part)

	description := monolith.Description{Parts:parts}
	args := make([]interface{}, 0)
	monolithInstance := monolith.Instance{
		Desc: description,
		Args: args,
	}

	addSequences := monolithInstance
	removeSequences := description

	monotoneConfig := toneburst.MonotoneConfig{
		AddSequences:    &addSequences,
		RemoveSequences: &removeSequences,
		SpeakFirst:      false,
	}

	return monotoneConfig
}

func createMonotoneClientConfigRandomEnumeratedItems() toneburst.MonotoneConfig {
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

	return monotoneConfig
}

func createMonotoneServerConfigRandomEnumeratedItems() toneburst.MonotoneConfig {
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

	return monotoneConfig
}

// Polish Tests
func TestConn(t *testing.T) {
	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		println("Polish server error: ", polishServerError)
		t.Fail()
	}

	serverConfig := ServerConfig{
		Toneburst: nil,
		Polish:    polishServerConfig,
	}

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
	polishClientConfig, polishClientError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientError != nil {
		println("Polish  client error: ", polishClientError)
		t.Fail()
		return
	}

	clientConfig := ClientConfig{
		Toneburst: nil,
		Polish:    polishClientConfig,
	}

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
	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		println("Polish server error: ", polishServerError)
		t.Fail()
	}

	polishClientConfig, polishClientConfigError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientConfigError != nil {
		println("Error creating silver client config: ", polishClientConfigError)
		t.Fail()
		return
	}

	monotoneClientConfig := createMonotoneClientConfigRandomEnumeratedItems()
	monotoneServerConfig := createMonotoneServerConfigRandomEnumeratedItems()
	//monotoneConfig := createMonotoneConfig()

	serverConfig := ServerConfig{
		Toneburst: monotoneServerConfig,
		Polish:    polishServerConfig,
	}

	clientConfig := ClientConfig{
		Toneburst: monotoneClientConfig,
		Polish:    polishClientConfig,
	}

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
