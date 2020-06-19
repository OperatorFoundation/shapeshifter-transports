package meekserver

import (
	"testing"
)

//if disableTLS is true, it doesnt require the cert and key
//func TestMeekServerListen(t *testing.T) {
//	meekserverTransport := NewMeekTransportServer(true, "", "", "state")
//	listener := meekserverTransport.Listen("127.0.0.1:80")
//	if listener == nil {
//		t.Fail()
//	}
//}

func TestMeekServerListen2(t *testing.T) {
	acmeEmail := "brandon@operatorfoundation.org"
	keyFileName := "operatorrss.com"
	meekserverTransport := NewMeekTransportServer(false, acmeEmail, keyFileName, "state")
	if meekserverTransport == nil {
		t.Fail()
		return
	}
	listener := meekserverTransport.Listen("127.0.0.1:8080")
	if listener == nil {
		t.Fail()
		return
	}
}