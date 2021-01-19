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

package replicant

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"github.com/OperatorFoundation/monolith-go/monolith"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v3/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v3/toneburst"
)

func InitializeGobRegistry() {
	monolith.InitializeGobRegistry()

	gob.Register(toneburst.MonotoneConfig{})
	gob.Register(toneburst.WhalesongConfig{})
	gob.Register(polish.SilverPolishClientConfig{})
	gob.Register(polish.SilverPolishServerConfig{})
}

func (config ClientConfig) Encode() (string, error) {
	InitializeGobRegistry()

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	marshalError := encoder.Encode(config)
	if marshalError != nil {
		return "", marshalError
	}

	encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())
	return encoded, nil
}

func DecodeClientConfig(encoded string) (*ClientConfig, error) {
	InitializeGobRegistry()

	decoded, base64Error := base64.StdEncoding.DecodeString(encoded)
	if base64Error != nil {
		return nil, base64Error
	}

	var buffer bytes.Buffer
	buffer.Write(decoded)

	decoder := gob.NewDecoder(&buffer)

	var config ClientConfig
	unmarshalError := decoder.Decode(&config)
	if unmarshalError != nil {
		return nil, unmarshalError
	}

	return &config, nil
}

func (config ServerConfig) Encode() (string, error) {
	InitializeGobRegistry()

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	marshalError := encoder.Encode(config)
	if marshalError != nil {
		return "", marshalError
	}

	encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())
	return encoded, nil
}

func DecodeServerConfig(encoded string) (*ServerConfig, error) {
	InitializeGobRegistry()

	decoded, base64Error := base64.StdEncoding.DecodeString(encoded)
	if base64Error != nil {
		return nil, base64Error
	}

	var buffer bytes.Buffer
	buffer.Write(decoded)

	decoder := gob.NewDecoder(&buffer)

	var config ServerConfig
	unmarshalError := decoder.Decode(&config)
	if unmarshalError != nil {
		return nil, unmarshalError
	}

	return &config, nil
}

