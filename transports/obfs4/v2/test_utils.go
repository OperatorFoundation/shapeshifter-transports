package obfs4

import (
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
		if !os.IsExist(directoryErr){
			return false
		}
	}
	serverConfig, confError := NewObfs4Server(fPath)
	if confError != nil {
		return false
	}
	//call listen on the server
	serverListener := serverConfig.Listen("127.0.0.1:1234")
	if serverListener == nil {
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
