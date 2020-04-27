package optimizer

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/meeklite/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow/v2"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	config := shadow.NewConfig("orange", "aes-128-ctr")
	listener := config.Listen("127.0.0.1:1235")
	go acceptConnections(listener)

	_ = obfs4.RunLocalObfs4Server("test")

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
	shadowTransport := shadow.Transport{Password: "orange", CipherName: "aes-128-ctr", Address: "127.0.0.1:1235"}
	_, err := shadowTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestMeekliteDial1(t *testing.T) {
	unparsedUrl := "https://d2zfqthxsdq309.cloudfront.net/"
	Url, parseErr := url.Parse(unparsedUrl)
	if parseErr != nil {
		t.Fail()
	}
	meekliteTransport := meeklite.Transport{Url: Url, Front: "a0.awsstatic.com", Address: "127.0.0.1:1235" }
	_, err := meekliteTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerMeekliteDial1(t *testing.T) {
	unparsedUrl := "https://d2zfqthxsdq309.cloudfront.net/"
	Url, parseErr := url.Parse(unparsedUrl)
	if parseErr != nil {
		t.Fail()
	}
	meekliteTransport := meeklite.Transport{Url: Url, Front: "a0.awsstatic.com", Address: "127.0.0.1:1235" }
	transports := []Transport{meekliteTransport}
	strategy := NewFirstStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestShadowDial2(t *testing.T) {
	shadowTransport := shadow.Transport{Password: "1234", CipherName: "CHACHA20-IETF-POLY1305", Address: "127.0.0.1:1235"}
	_, err := shadowTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerShadowDial1(t *testing.T) {
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{&shadowTransport}
	strategy := NewFirstStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerShadowDial2(t *testing.T) {
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{&shadowTransport}
	strategy := NewFirstStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)
	_, err:= optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestObfs4Transport_Dial1(t *testing.T) {

	obfs4Transport, transportErr := obfs4.RunObfs4Client()
	if transportErr != nil {
		t.Fail()
		return
	}
	_, err := obfs4Transport.Dial("127.0.0.1:1234")
	if err != nil {
		t.Fail()
	}
}

func TestObfs4Transport_Dial2(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.OptimizerTransport{
		CertString: "60RNHBMRrf+aOSPzSj8bD4ASGyyPl0mkaOUAQsAYljSkFB0G8B8m9fGvGJCpOxwoXS1baA",
		IatMode:    0,
		Address:    "127.0.0.1:1234",
		Dialer:     dialer}
	_, err := obfs4Transport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerObfs4Transport_Dial1(t *testing.T) {
	dialer := proxy.Direct
	certstring, certError := getObfs4CertString()
	if certError != nil {
		t.Fail()
		return
	}
	obfs4Transport := obfs4.OptimizerTransport{
		CertString: *certstring,
		IatMode:    0,
		Address:    "127.0.0.1:1234",
		Dialer:     dialer}
	transports := []Transport{obfs4Transport}
	strategy := NewFirstStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerObfs4Transport_Dial2(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.OptimizerTransport{
		CertString: "60RNHBMRrf+aOSPzSj8bD4ASGyyPl0mkaOUAQsAYljSkFB0G8B8m9fGvGJCpOxwoXS1baA",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer}
	transports := []Transport{obfs4Transport}
	strategy := NewFirstStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerTransportFirstDial(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.OptimizerTransport{
		CertString: "60RNHBMRrf+aOSPzSj8bD4ASGyyPl0mkaOUAQsAYljSkFB0G8B8m9fGvGJCpOxwoXS1baA",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer}
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{obfs4Transport, &shadowTransport}
	strategy := NewFirstStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)
	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}

func TestOptimizerTransportRandomDial(t *testing.T) {
	dialer := proxy.Direct
	obfs4Transport := obfs4.OptimizerTransport{
		CertString: "60RNHBMRrf+aOSPzSj8bD4ASGyyPl0mkaOUAQsAYljSkFB0G8B8m9fGvGJCpOxwoXS1baA",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,
	}
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{obfs4Transport, &shadowTransport}
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
	obfs4Transport := obfs4.OptimizerTransport{
		CertString: "60RNHBMRrf+aOSPzSj8bD4ASGyyPl0mkaOUAQsAYljSkFB0G8B8m9fGvGJCpOxwoXS1baA",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,
}
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{obfs4Transport, &shadowTransport}
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
	obfs4Transport := obfs4.OptimizerTransport{
		CertString: "60RNHBMRrf+aOSPzSj8bD4ASGyyPl0mkaOUAQsAYljSkFB0G8B8m9fGvGJCpOxwoXS1baA",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,
}
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{obfs4Transport, &shadowTransport}
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
	obfs4Transport := obfs4.OptimizerTransport{
		CertString: "60RNHBMRrf+aOSPzSj8bD4ASGyyPl0mkaOUAQsAYljSkFB0G8B8m9fGvGJCpOxwoXS1baA",
		IatMode:    0,
		Address:    "77.81.104.251:443",
		Dialer:     dialer,
	}
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{obfs4Transport, &shadowTransport}
	strategy := NewMinimizeDialDuration(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)

	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}

func getObfs4CertString() (*string, error){
	fPath := path.Join("/Users/bluesaxorcist/stateDir", "obfs4_bridgeline.txt")
	bytes, fileError := ioutil.ReadFile(fPath)
	if fileError != nil {
		return nil, fileError
	}
	//print(bytes)
	byteString := string(bytes)
	//print(byteString)
	lines := strings.Split(byteString, "\n")
	//print(lines)
	bridgeLine := lines[len(lines)-2]
	//println(bridgeLine)
	bridgeParts1 := strings.Split(bridgeLine, " ")
	bridgePart := bridgeParts1[5]
	certstring := bridgePart[5:]

	return &certstring, nil
}