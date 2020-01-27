package replicant

import (
	"bytes"
	"fmt"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/polish"
	"net"
	"testing"
)

// Polish Tests

func TestConn(t *testing.T) {
	message := "Hi there!\n"

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		println("Polish server error: ", polishServerError)
		t.Fail()
	}

	go func() {

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
		readBuffer := make([]byte, 1024)
		go client.Read(readBuffer)
		writeBytes := []byte{0x0A, 0x11, 0xB0, 0xB1}
		writeCount, writeError := client.Write(writeBytes)
		if writeError != nil {
			println("Write error from client: ", writeError)
		}

		println("Write byte count = ", writeCount)

		//conn, err := net.Dial("tcp", ":3000")
		//if err != nil {
		//	t.Fatal(err)
		//}
		defer client.Close()


		if _, err := fmt.Fprintf(client, message); err != nil {
			println("error", err)
			t.Fatal(err)
		}
	}()

	// Run the server

	serverConfig := ServerConfig{
		Toneburst: nil,
		Polish:    polishServerConfig,
	}

	listener := serverConfig.Listen("127.0.0.1:7777")
	//listener, err := net.Listen("tcp", ":3000")
	//if err != nil {
	//	t.Fatal(err)
	//}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			println("Listener accept func error: ", err)
			return
		}
		defer conn.Close()

		//buf, err := ioutil.ReadAll(conn)
		//if err != nil {
		//	println("read all buffer error: ", err)
		//	t.Fatal(err)
		//	return
		//}

		println("received an incoming connection")
		// Make a buffer to hold incoming data.
		buf := make([]byte, 1024)
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
		}

		println("regLen: ", reqLen)
		println("readBuffer: ", buf)
		// Send a response back to person contacting us.
		conn.Write([]byte("Message received."))


		fmt.Println(string(buf[:]))
		if msg := string(buf[:]); msg != message {
			t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", msg, message)
		}
		return // Done
	}

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

func TestServerWithSilverDialogue(t *testing.T) {
	// Run the server
	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		t.Fail()
		return
	}
	serverConfig := ServerConfig{
		Toneburst: nil,
		Polish:    polishServerConfig,
	}

	listener := serverConfig.Listen("127.0.0.1:7777")

	go serverStuff(listener)
	clientStuff(polishServerConfig)
}

func serverStuff(listener net.Listener) {
	for {
		// Listen for an incoming connection.
		server, serverError := listener.Accept()
		if serverError != nil {
			println("server error: ", serverError)
			return
		}

		go handleRequest(server)
	}
}

func clientStuff(polishServerConfig *polish.SilverPolishServerConfig) {
	// Run the client
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

func handleRequest(conn net.Conn) {
	println("received an incoming connection")
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	println("regLen: ", reqLen)
	println("readBuffer: ", buf)
	// Send a response back to person contacting us.
	conn.Write([]byte("Message received."))
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
