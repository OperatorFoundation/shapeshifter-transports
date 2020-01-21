package replicant

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/toneburst"
	"testing"
)

func TestEncodeClientConfig(t *testing.T) {
	toneburstConfig := toneburst.WhalesongConfig{
		AddSequences:    []toneburst.Sequence{},
		RemoveSequences: []toneburst.Sequence{},
	}

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		t.Fail()
		return
	}
	polishClientConfig, polishClientError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientError != nil {
		t.Fail()
		return
	}

	config := ClientConfig{
		Toneburst: toneburstConfig,
		Polish:    polishClientConfig,
	}

	_, marshalError := config.Encode()
	if marshalError != nil {
		t.Fail()
		return
	}
}

func TestDecodeClientConfig(t *testing.T) {
	toneburstConfig := toneburst.WhalesongConfig{
		AddSequences:    []toneburst.Sequence{},
		RemoveSequences: []toneburst.Sequence{},
	}

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		t.Fail()
		return
	}
	polishClientConfig, polishClientError := polish.NewSilverClientConfig(polishServerConfig)
	if polishClientError != nil {
		t.Fail()
		return
	}

	config := ClientConfig{
		Toneburst: toneburstConfig,
		Polish:    polishClientConfig,
	}

	result, marshalError := config.Encode()
	if marshalError != nil {
		t.Fail()
		return
	}

	_, unmarshalError := DecodeClientConfig(result)
	if unmarshalError != nil {
		t.Fail()
		return
	}
}

func TestEncodeServerConfig(t *testing.T) {
	toneburstConfig := toneburst.WhalesongConfig{
		AddSequences:    []toneburst.Sequence{},
		RemoveSequences: []toneburst.Sequence{},
	}

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		t.Fail()
		return
	}

	config := ServerConfig{
		Toneburst: toneburstConfig,
		Polish:    polishServerConfig,
	}

	_, marshalError := config.Encode()
	if marshalError != nil {
		t.Fail()
		return
	}
}

func TestDecodeServerConfig(t *testing.T) {
	toneburstConfig := toneburst.WhalesongConfig{
		AddSequences:    []toneburst.Sequence{},
		RemoveSequences: []toneburst.Sequence{},
	}

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		t.Fail()
		return
	}

	config := ServerConfig{
		Toneburst: toneburstConfig,
		Polish:    polishServerConfig,
	}

	result, marshalError := config.Encode()
	if marshalError != nil {
		t.Fail()
		return
	}

	_, unmarshalError := DecodeServerConfig(result)
	if unmarshalError != nil {
		t.Fail()
		return
	}
}