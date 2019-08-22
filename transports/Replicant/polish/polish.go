package polish

import (
	"fmt"
	"net"
)

type PolishConnection interface {
	Handshake(conn net.Conn) error
	Polish(input []byte) []byte
	Unpolish(input []byte) []byte
}

type PolishServer interface {
	NewConnection(net.Conn) PolishConnection
}

func NewClient(config Config) interface{PolishConnection} {
	switch config.selector {
	case "silver":
		if config.silver == nil {
			fmt.Println("Error, silver config missing")
			return nil
		} else if !config.silver.clientOrServer {
			fmt.Println("Error, tried to initialize client, but config was not client config")
			return nil
		} else if config.silver.clientConfig == nil {
			fmt.Println("Error, tried to initialize client, but client config was missing")
			return nil
		} else {
			return NewSilverClient(*config.silver.clientConfig)
		}
	default:
		return nil
	}
}

func NewServer(config Config) interface{PolishServer} {
	switch config.selector {
	case "silver":
		if config.silver == nil {
			fmt.Println("Error, silver config missing")
			return nil
		} else if config.silver.clientOrServer {
			fmt.Println("Error, tried to initialize server, but config was not server config")
			return nil
		} else if config.silver.serverConfig == nil {
			fmt.Println("Error, tried to initialize server, but server config was missing")
			return nil
		} else {
			return NewSilverServer(*config.silver.serverConfig)
		}
	default:
		return nil
	}
}
