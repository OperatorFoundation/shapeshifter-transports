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

