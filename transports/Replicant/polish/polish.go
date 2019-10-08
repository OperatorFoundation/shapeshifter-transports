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

func NewClient(config *Config) PolishConnection {
	if config == nil {
		return nil
	}

	switch config.Selector {
	case "Silver":
		if config.Silver == nil {
			fmt.Println("Error, Silver config missing")
			return nil
		} else if !config.Silver.ClientOrServer {
			fmt.Println("Error, tried to initialize client, but config was not client config")
			return nil
		} else if config.Silver.ClientConfig == nil {
			fmt.Println("Error, tried to initialize client, but client config was missing")
			return nil
		} else {
			return NewSilverClient(*config.Silver.ClientConfig)
		}
	default:
		return nil
	}
}

func NewServer(config Config) interface{PolishServer} {
	switch config.Selector {
	case "Silver":
		if config.Silver == nil {
			fmt.Println("Error, Silver config missing")
			return nil
		} else if config.Silver.ClientOrServer {
			fmt.Println("Error, tried to initialize server, but config was not server config")
			return nil
		} else if config.Silver.ServerConfig == nil {
			fmt.Println("Error, tried to initialize server, but server config was missing")
			return nil
		} else {
			return NewSilverServer(*config.Silver.ServerConfig)
		}
	default:
		return nil
	}
}
