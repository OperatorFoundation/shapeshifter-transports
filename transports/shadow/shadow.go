/*
 * Copyright (c) 2020, Operator Foundation
 *
 */

// Package shadow provides a PT 2.1 Go API wrapper around the connections used by Shadowsocks
package shadow

import (
	"log"
	"net"

	shadowsocks "github.com/shadowsocks/go-shadowsocks2/core"
)

type Config struct {
	Password   string `json:"password"`
	CipherName string `json:"cipherName"`
}

type Transport struct {
	Password   string
	CipherName string
	Address    string
}

func NewConfig(password string, cipherName string) Config {
	return Config{
		Password:   password,
		CipherName: cipherName,
	}
}

func NewTransport(password string, cipherName string, address string) Transport {
	return Transport{
		Password:   password,
		CipherName: cipherName,
		Address:    address,
	}
}

func (config Config) Listen(address string) (net.Listener, error) {
	cipher, err := shadowsocks.PickCipher(config.CipherName, nil, config.Password)
	if err != nil {
		log.Fatal("Failed generating ciphers:", err)
	}

	return shadowsocks.Listen("tcp", address, cipher)
}

func (config Config) Dial(address string) (net.Conn, error) {
	cipher, err := shadowsocks.PickCipher(config.CipherName, nil, config.Password)
	if err != nil {
		log.Fatal("Failed generating ciphers:", err)
	}

	conn, err := shadowsocks.Dial("tcp", address, cipher)
	if err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

//begin code added from optimizer
// Create outgoing transport connection
func (transport *Transport) Dial() (net.Conn, error) {
	cipher, err := shadowsocks.PickCipher(transport.CipherName, nil, transport.Password)
	if err != nil {
		log.Fatal("Failed generating ciphers:", err)
	}

	return shadowsocks.Dial("tcp", transport.Address, cipher)
}
//end code added from optimizer
