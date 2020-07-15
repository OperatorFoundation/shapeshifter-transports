package meekserver

import (
	"golang.org/x/crypto/acme/autocert"
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
func TestMeekServerListen(t *testing.T) {
	acmeEmail := "brandon@operatorfoundation.org"
	keyFileName := "operatorrss.com"
	meekserverTransport := NewMeekTransportServer(false, acmeEmail, keyFileName, "state")
	if meekserverTransport == nil {
		t.Fail()
		return
	}
	_, listenErr := meekserverTransport.Listen("127.0.0.1:8080")
	if listenErr != nil {
		t.Fail()
		return
	}
}

func TestMeekServerFactoryListen(t *testing.T) {
	cert:= autocert.Manager{}
	meekserverTransport := New(false, &cert,"127.0.0.1:8080" )
	//TODO why does it try to convert nil into type Transport? Is a nil check needed here?
	//if meekserverTransport == nil {
	//	t.Fail()
	//	return
	//}
	_, listenErr := meekserverTransport.Listen()
	if listenErr != nil {
		t.Fail()
		return
	}
}