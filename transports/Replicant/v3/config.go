package replicant

import (
	"encoding/json"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v3/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v3/toneburst"
)

type ClientConfig struct {
	Toneburst toneburst.Config
	Polish    polish.ClientConfig
}

type ServerConfig struct {
	Toneburst toneburst.Config
	Polish    polish.ServerConfig
}

func (config ServerConfig) Marshal() (string, error) {

	type ServerJSONInnerConfig struct {
		Config string `json:"config"`
	}

	type ServerJSONOuterConfig struct {
		Replicant ServerJSONInnerConfig
	}

	configString, configStringError := config.Encode()
	if configStringError != nil {
		return "", configStringError
	}

	innerConfig := ServerJSONInnerConfig{Config:configString}
	outerConfig := ServerJSONOuterConfig{Replicant:innerConfig}

	configBytes, marshalError := json.Marshal(outerConfig)
	if marshalError != nil {
		return "", marshalError
	}

	return string(configBytes), nil
}

func (config ClientConfig) Marshal() (string, error) {

	type ClientJSONConfig struct {
		Config string `json:"config"`
	}

	configString, configStringError := config.Encode()
	if configStringError != nil {
		return "", configStringError
	}

	clientConfig := ClientJSONConfig{Config:configString}

	configBytes, marshalError := json.Marshal(clientConfig)
	if marshalError != nil {
		return "", marshalError
	}

	return string(configBytes), nil
}