package shadow

import (
	"golang.org/x/net/proxy"
	"testing"
)

const data = "test"

func TestShadow(t *testing.T) {
	//create a server
	shadowServer := NewShadowServer("password", "aes-128-ctr")
	//create client
	shadowClient := NewShadowClient("password", "aes-128-ctr", proxy.Direct)
	//call listen on the server
	serverListener := shadowServer.Listen("127.0.0.1:1234")
	//Create Server connection and format it for concurrency
	go func() {
		//create server buffer
		serverBuffer := make([]byte, 4)
		//create serverConn
		serverConn, acceptErr := serverListener.Accept()
		if acceptErr != nil {
			t.Fail()
			return
		}
		//read on server side
		_, serverReadErr := serverConn.Read(serverBuffer)
		if serverReadErr != nil {
			t.Fail()
			return
		}
		//write data from serverConn for client to read
		_, serverWriteErr := serverConn.Write([]byte(data))
		if serverWriteErr != nil {
			t.Fail()
			return
		}
	}()

	//create client buffer
	clientBuffer := make([]byte, 4)
	//call dial on client and check error
	clientConn, dialErr := shadowClient.Dial("127.0.0.1:1234")
	if dialErr != nil {
		t.Fail()
		return
	}

	//write data from clientConn for server to read
	_, clientWriteErr := clientConn.Write([]byte(data))
	if clientWriteErr != nil {
		t.Fail()
		return
	}

	//read on client side
	_, clientReadErr := clientConn.Read(clientBuffer)
	if clientReadErr != nil {
		t.Fail()
		return
	}
}
