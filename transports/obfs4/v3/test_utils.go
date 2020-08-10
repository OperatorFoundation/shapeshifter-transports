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

package obfs4

import (
	"github.com/kataras/golog"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
)

//RunLocalObfs4Server runs the server side in the background for the test
func RunLocalObfs4Server(data string) bool {
	//create a server
	usr, userError := user.Current()
	if userError != nil {
		return false
	}
	home := usr.HomeDir
	var fPath string
	if runtime.GOOS == "darwin" {
		fPath = path.Join(home, "shapeshifter-transports/stateDir")
	} else {
		fPath = path.Join(home, "gopath/src/github.com/OperatorFoundation/shapeshifter-transports/stateDir")
	}
	directoryErr := os.Mkdir(fPath, 0775)
	if directoryErr != nil {
		if !os.IsExist(directoryErr) {
			return false
		}
	}
	serverConfig, confError := NewObfs4Server(fPath)
	if confError != nil {
		return false
	}
	//call listen on the server
	serverListener, listenErr := serverConfig.Listen("127.0.0.1:1234")
	if listenErr != nil {
		return false
	}
	//Create Server connection and format it for concurrency
	go func() {
		//create server buffer
		serverBuffer := make([]byte, 4)

		for {
			//create serverConn
			serverConn, acceptErr := serverListener.Accept()
			if acceptErr != nil {
				return
			}

			go func() {
				//read on server side
				_, serverReadErr := serverConn.Read(serverBuffer)
				if serverReadErr != nil {
					return
				}

				//write data from serverConn for client to read
				_, serverWriteErr := serverConn.Write([]byte(data))
				if serverWriteErr != nil {
					return
				}
			}()
		}
	}()
	return true
}

func RunLocalObfs4ServerFactory(data string) bool {
	//create a server
	MakeLog()
	usr, userError := user.Current()
	if userError != nil {
		return false
	}
	home := usr.HomeDir
	var fPath string
	if runtime.GOOS == "darwin" {
		fPath = path.Join(home, "shapeshifter-transports/stateDir")
	} else {
		fPath = path.Join(home, "gopath/src/github.com/OperatorFoundation/shapeshifter-transports/stateDir")
	}
	directoryErr := os.Mkdir(fPath, 0775)
	if directoryErr != nil {
		if !os.IsExist(directoryErr) {
			return false
		}
	}
	serverConfig, confError := NewServer(fPath, "127.0.0.1:2234")
	if confError != nil {
		return false
	}
	//call listen on the server
	serverListener, listenErr := serverConfig.Listen()
	if listenErr != nil {
		return false
	}
	//Create Server connection and format it for concurrency
	go func() {
		//create server buffer
		serverBuffer := make([]byte, 4)

		for {
			//create serverConn
			serverConn, acceptErr := serverListener.Accept()
			if acceptErr != nil {
				return
			}

			go func() {
				//read on server side
				_, serverReadErr := serverConn.Read(serverBuffer)
				if serverReadErr != nil {
					return
				}

				//write data from serverConn for client to read
				_, serverWriteErr := serverConn.Write([]byte(data))
				if serverWriteErr != nil {
					return
				}
			}()
		}
	}()
	return true
}

//RunObfs4Client runs the client side in the background for the test
func RunObfs4Client() (*Transport, error) {
	usr, userError := user.Current()
	if userError != nil {
		return nil, userError
	}
	home := usr.HomeDir
	var fPath string
	if runtime.GOOS == "darwin" {
		fPath = path.Join(home, "shapeshifter-transports/stateDir/obfs4_bridgeline.txt")
	} else {
		fPath = path.Join(home, "gopath/src/github.com/OperatorFoundation/shapeshifter-transports/stateDir/obfs4_bridgeline.txt")
	}
	bytes, fileError := ioutil.ReadFile(fPath)
	if fileError != nil {
		return nil, fileError
	}
	//print(bytes)
	byteString := string(bytes)
	//print(byteString)
	lines := strings.Split(byteString, "\n")
	//print(lines)
	bridgeLine := lines[len(lines)-2]
	//println(bridgeLine)
	bridgeParts1 := strings.Split(bridgeLine, " ")
	bridgePart := bridgeParts1[5]
	certstring := bridgePart[5:]
	//println(certstring)
	clientConfig, confError := NewObfs4Client(certstring, 0, nil)
	return clientConfig, confError
}

//RunObfs4Client runs the client side in the background for the test
func RunObfs4ClientFactory() (*TransportClient, error) {
	MakeLog()
	usr, userError := user.Current()
	if userError != nil {
		return nil, userError
	}
	home := usr.HomeDir
	var fPath string
	if runtime.GOOS == "darwin" {
		fPath = path.Join(home, "shapeshifter-transports/stateDir/obfs4_bridgeline.txt")
	} else {
		fPath = path.Join(home, "gopath/src/github.com/OperatorFoundation/shapeshifter-transports/stateDir/obfs4_bridgeline.txt")
	}
	bytes, fileError := ioutil.ReadFile(fPath)
	if fileError != nil {
		return nil, fileError
	}
	//print(bytes)
	byteString := string(bytes)
	//print(byteString)
	lines := strings.Split(byteString, "\n")
	//print(lines)
	bridgeLine := lines[len(lines)-2]
	//println(bridgeLine)
	bridgeParts1 := strings.Split(bridgeLine, " ")
	bridgePart := bridgeParts1[5]
	certstring := bridgePart[5:]
	//println(certstring)
	clientConfig, confError := NewClient(certstring, 0, "127.0.0.1:2234", proxy.Direct)
	return &clientConfig, confError
}

func MakeLog() {
	golog.SetLevel("debug")
	golog.SetOutput(os.Stderr)
}
