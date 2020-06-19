package replicant

import (
	"encoding/json"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2/toneburst"
	"io/ioutil"
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

type ReplicantJSONConfig struct {
	Config string `json:"config"`
}


func TestSaveServerConfigPolish(t *testing.T) {

	polishServerConfig, polishServerError := polish.NewSilverServerConfig()
	if polishServerError != nil {
		t.Fail()
		return
	}

	config := ServerConfig{
		Toneburst: nil,
		Polish:    polishServerConfig,
	}

	serverConfigString, marshalError := config.Encode()
	if marshalError != nil {
		t.Fail()
		return
	}

	// Create a struct
	rServerJSON := ReplicantJSONConfig{Config: serverConfigString}

	// Save it to a json file
	jsonString, jsonError := json.Marshal(rServerJSON)
	if jsonError != nil {
		t.Fail()
		return
	}

	_ = ioutil.WriteFile("/Users/mafalda/Documents/Operator/replicant_server.json", jsonString, 0644)
}

func TestSaveClientConfigPolish(t *testing.T) {
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
		Toneburst: nil,
		Polish:    polishClientConfig,
	}

	clientConfigString, marshalError := config.Encode()
	if marshalError != nil {
		t.Fail()
		return
	}

	rClientJSON := ReplicantJSONConfig{clientConfigString}
	jsonString, jsonError := json.Marshal(rClientJSON)
	if jsonError != nil {
		t.Fail()
		return
	}

	_ = ioutil.WriteFile("/Users/mafalda/Documents/Operator/replicant_client.json", jsonString, 0644)
}