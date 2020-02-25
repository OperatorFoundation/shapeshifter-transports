/*
	MIT License

	Copyright (c) 2020 Operator Foundation

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

package shadow

import (
	"testing"
)

const data = "test"

func TestShadow(t *testing.T) {
	//create a server
	config := NewConfig("password", "aes-128-ctr")

	//call listen on the server
	serverListener := config.Listen("127.0.0.1:1234")
	if serverListener != nil {
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
	clientConn, dialErr := config.Dial("127.0.0.1:1234")
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
