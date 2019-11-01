package Optimizer

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/meeklite"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow"
	"golang.org/x/net/proxy"
	"net"
	"net/url"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	transport := shadow.NewShadowServer("orange", "aes-128-ctr")
	listener := transport.Listen("127.0.0.1:1234")
	go acceptConnections(listener)

	os.Exit(m.Run())
}

func acceptConnections(listener net.Listener) {
	for {
		_, err := listener.Accept()
		if err != nil {
			return
		}
	}
}

func TestShadowDial1(t *testing.T) {
	shadowTransport := shadow.Transport{"orange", "aes-128-ctr", "127.0.0.1:1234"}
	_, err := shadowTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestMeekliteDial1(t *testing.T) {
	unparsedUrl := "https://d2zfqthxsdq309.cloudfront.net/"
	Url, _ := url.Parse(unparsedUrl)
	meekliteTransport := meeklite.Transport{Url, "a0.awsstatic.com", "127.0.0.1:1234" }
	_, err := meekliteTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerMeekliteDial1(t *testing.T) {
	unparsedUrl := "https://d2zfqthxsdq309.cloudfront.net/"
	Url, _ := url.Parse(unparsedUrl)
	meekliteTransport := meeklite.Transport{Url, "a0.awsstatic.com", "127.0.0.1:1234" }
	transports := []Transport{meekliteTransport}
	strategy := FirstStrategy{}
	optimizerTransport := NewOptimizerClient(transports, &strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestShadowDial2(t *testing.T) {
	shadowTransport := shadow.Transport{"banana", "aes-192-ctr", "127.0.0.1:1234"}
	_, err := shadowTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerShadowDial1(t *testing.T) {
	shadowTransport := shadow.Transport{"orange", "aes-128-ctr", "127.0.0.1:1234"}
	transports := []Transport{shadowTransport}
	strategy := FirstStrategy{}
	optimizerTransport := NewOptimizerClient(transports, &strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerShadowDial2(t *testing.T) {
	shadowTransport := shadow.Transport{"banana", "aes-192-ctr", "127.0.0.1:1234"}
	transports := []Transport{shadowTransport}
	strategy := FirstStrategy{}
	optimizerTransport := NewOptimizerClient(transports, &strategy)
	_, err:= optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestObfs4Transport_Dial1(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "UsuF7oN4KNKviZP54JOyTCoCphrdM5gwZK4vT8GnCAcmqLUJEJxyw1dpko9a/ii6He4iZg",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,}
	_, err := obfs4Transport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestObfs4Transport_Dial2(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "UsuF7oN4KNKviZP54JOyTCoCphrdM5gwZK4vT8GnCAcmqLUJEJxyw1dpko9a/ii6He4iZg",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,}
	_, err := obfs4Transport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerObfs4Transport_Dial1(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "UsuF7oN4KNKviZP54JOyTCoCphrdM5gwZK4vT8GnCAcmqLUJEJxyw1dpko9a/ii6He4iZg",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,}
	transports := []Transport{obfs4Transport}
	strategy := FirstStrategy{}
	optimizerTransport := NewOptimizerClient(transports, &strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerObfs4Transport_Dial2(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "UsuF7oN4KNKviZP54JOyTCoCphrdM5gwZK4vT8GnCAcmqLUJEJxyw1dpko9a/ii6He4iZg",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,}
	transports := []Transport{obfs4Transport}
	strategy := FirstStrategy{}
	optimizerTransport := NewOptimizerClient(transports, &strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerTransportFirstDial(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "UsuF7oN4KNKviZP54JOyTCoCphrdM5gwZK4vT8GnCAcmqLUJEJxyw1dpko9a/ii6He4iZg",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,}
	shadowTransport := shadow.Transport{"orange", "aes-128-ctr", "127.0.0.1:1234"}
	transports := []Transport{obfs4Transport, shadowTransport}
	optimizerTransport := NewOptimizerClient(transports, &FirstStrategy{})
	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}

func TestOptimizerTransportRandomDial(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "UsuF7oN4KNKviZP54JOyTCoCphrdM5gwZK4vT8GnCAcmqLUJEJxyw1dpko9a/ii6He4iZg",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,
	}
	shadowTransport := shadow.Transport{"orange", "aes-128-ctr", "127.0.0.1:1234"}
	transports := []Transport{obfs4Transport, shadowTransport}
	optimizerTransport := NewOptimizerClient(transports, &RandomStrategy{})

	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}

func TestOptimizerTransportRotateDial(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "UsuF7oN4KNKviZP54JOyTCoCphrdM5gwZK4vT8GnCAcmqLUJEJxyw1dpko9a/ii6He4iZg",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,
}
shadowTransport := shadow.Transport{"orange", "aes-128-ctr", "127.0.0.1:1234"}
	transports := []Transport{obfs4Transport, shadowTransport}
	optimizerTransport := NewOptimizerClient(transports, &RotateStrategy{})

	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}

func TestOptimizerTransportTrackDial(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "UsuF7oN4KNKviZP54JOyTCoCphrdM5gwZK4vT8GnCAcmqLUJEJxyw1dpko9a/ii6He4iZg",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,
}
shadowTransport := shadow.Transport{"orange", "aes-128-ctr", "127.0.0.1:1234"}
	transports := []Transport{obfs4Transport, shadowTransport}
	optimizerTransport := NewOptimizerClient(transports, NewTrackStrategy(transports))

	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}

func TestOptimizerTransportminimizeDialDurationDial(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "UsuF7oN4KNKviZP54JOyTCoCphrdM5gwZK4vT8GnCAcmqLUJEJxyw1dpko9a/ii6He4iZg",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,
	}
	shadowTransport := shadow.Transport{"orange", "aes-128-ctr", "127.0.0.1:1234"}
	transports := []Transport{obfs4Transport, shadowTransport}
	strategy := NewMinimizeDialDuration(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)

	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}