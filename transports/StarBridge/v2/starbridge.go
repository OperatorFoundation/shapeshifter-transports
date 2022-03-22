package StarBridge

import (
	replicant "github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v3"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v3/polish"
	"io/ioutil"
	"net"
)

func NewClientConnection(conn net.Conn) (*replicant.Connection, error) {
	config := getClientConfig()
	return replicant.NewClientConnection(conn, config)
}

func NewServerConnection(conn net.Conn) (*replicant.Connection, error) {
	config := getServerConfig()
	return replicant.NewServerConnection(conn, config)
}

func NewReplicantClientConnectionState() (*replicant.ConnectionState, error) {
	config := getClientConfig()
	return replicant.NewReplicantClientConnectionState(config)
}

func NewReplicantServerConnectionState(polishServer polish.Server, conn net.Conn) (*replicant.ConnectionState, error) {
	config := getServerConfig()
	return replicant.NewReplicantServerConnectionState(config, polishServer, conn)
}

func getClientConfig() replicant.ClientConfig {
	clientConfigBytes, clientReadFileError := ioutil.ReadFile("StarBridgeClientConfig.txt")
	if clientReadFileError != nil {

	}

	clientConfig, clientDecodeError := replicant.DecodeClientConfig(string(clientConfigBytes))
	if clientDecodeError != nil {

	}

	return *clientConfig
}

func getServerConfig() replicant.ServerConfig {
	serverConfigBytes, serverReadFileError := ioutil.ReadFile("StarBridgeServerConfig.txt")
	if serverReadFileError != nil {

	}

	serverConfig, serverDecodeError := replicant.DecodeServerConfig(string(serverConfigBytes))
	if serverDecodeError != nil {
	}

	return *serverConfig
}
