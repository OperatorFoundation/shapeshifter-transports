package polish

import "testing"

func TestNewSilverServerConfig(t *testing.T) {
	_, configError := NewSilverServerConfig()
	if configError != nil {
		t.Fail()
		return
	}
}

func TestNewSilverClientConfig(t *testing.T) {
	serverConfig, serverConfigError := NewSilverServerConfig()
	if serverConfigError != nil {
		t.Fail()
		return
	}

	_, clientConfigError := NewSilverClientConfig(serverConfig)
	if clientConfigError != nil {
		t.Fail()
		return
	}
}