package Optimizer

import (
	"testing"
)

func TestShadowDial(t *testing.T) {
	shadowTransport := ShadowTransport{"orange", "aes-128-ctr", "127.0.0.1:1234"}
	conn := shadowTransport.Dial()
	if conn == nil {
		t.Fail()
	}
}

func TestOptimizerShadowDial (t *testing.T) {
	shadowTransport := ShadowTransport{"orange", "aes-128-ctr", "127.0.0.1:1234"}
	transports := []Transport{shadowTransport}
	optimizerTransport := optimizerTransport{transports}
	conn := optimizerTransport.Dial()
	if conn == nil {
		t.Fail()
	}
}

func TestObfs4Transport_Dial(t *testing.T) {
	Obfs4Transport := Obfs4Transport{"idk lol", "i have no clue"}
	conn := Obfs4Transport.Dial()
	if conn == nil {
		t.Fail()
	}
}

func TestOptimizerObfs4Transport_Dial(t *testing.T) {
	Obfs4Transport := Obfs4Transport{"idk lol", "i have no clue"}
	transports := []Transport{Obfs4Transport}
	optimizerTransport := optimizerTransport{transports}
	conn := optimizerTransport.Dial()
	if conn == nil {
		t.Fail()
	}
}