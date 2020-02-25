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

func (config Config) Listen(address string) net.Listener {
	cipher, err := shadowsocks.PickCipher(config.CipherName, nil, config.Password)
	if err != nil {
		log.Fatal("Failed generating ciphers:", err)
		return nil
	}

	listener, listenerErr := shadowsocks.Listen("tcp", address, cipher)
	if listenerErr != nil {
		log.Fatal("Failed to start listener:", listenerErr)
		return nil
	}
	return listener
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
