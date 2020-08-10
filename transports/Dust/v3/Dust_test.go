package Dust

/*
 * Copyright (c) 2019, Operator Foundation
 *
 */

// Package Dust provides a PT 2.1 Go API wrapper around the Dust transport


import (
	"testing"
)

const data = "test"

func TestDust(t *testing.T) {
	//create a server
	serverConfig := dustServer{
		serverPrivkey: nil,
		dialer:        nil,
	}
	clientConfig := dustClient{
		serverPubkey: nil,
		dialer:       nil,
	}

	//call listen on the server
	serverListener := serverConfig.Listen("127.0.0.1:1234")
	if serverListener == nil {
		t.Fail()
		return
	}

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
	clientConn, dialErr := clientConfig.Dial("127.0.0.1:1234")
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
