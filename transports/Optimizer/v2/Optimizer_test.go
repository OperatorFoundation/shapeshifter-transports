package optimizer

import (
	"github.com/OperatorFoundation/shapeshifter-transports/transports/meeklite/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs2/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow/v2"

	"golang.org/x/net/proxy"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
	"testing"
)

const data = "test"

func TestMain(m *testing.M) {
	config := shadow.NewConfig("1234", "CHACHA20-IETF-POLY1305")
	listener := config.Listen("127.0.0.1:1235")
	go acceptConnections(listener)

	_ = obfs4.RunLocalObfs4Server("test")

	RunLocalObfs2Server()

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

func TestMeekliteDial(t *testing.T) {
	unparsedURL := "https://d2zfqthxsdq309.cloudfront.net/"
	URL, parseErr := url.Parse(unparsedURL)
	if parseErr != nil {
		t.Fail()
	}
	meekliteTransport := meeklite.Transport{URL: URL, Front: "a0.awsstatic.com", Address: "127.0.0.1:1235" }
	_, err := meekliteTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerMeekliteDial(t *testing.T) {
	unparsedURL := "https://d2zfqthxsdq309.cloudfront.net/"
	URL, parseErr := url.Parse(unparsedURL)
	if parseErr != nil {
		t.Fail()
	}
	meekliteTransport := meeklite.Transport{URL: URL, Front: "a0.awsstatic.com", Address: "127.0.0.1:1235" }
	transports := []Transport{meekliteTransport}
	strategy := NewFirstStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestShadowDial(t *testing.T) {
	shadowTransport := shadow.Transport{Password: "1234", CipherName: "CHACHA20-IETF-POLY1305", Address: "127.0.0.1:1235"}
	_, err := shadowTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerShadowDial(t *testing.T) {
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{&shadowTransport}
	strategy := NewFirstStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestOptimizerObfs2Dial(t *testing.T) {
	obfs2Transport := obfs2.New("127.0.0.1:1237", proxy.Direct)
	transports := []Transport{obfs2Transport}
	strategy := NewFirstStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)
	_, err := optimizerTransport.Dial()
	if err != nil {
		t.Fail()
	}
}

func TestObfs4Transport_Dial(t *testing.T) {
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

func TestOptimizerObfs4Transport_Dial(t *testing.T) {
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

func TestOptimizerTransportFirstDial(t *testing.T) {
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
	certstring, certError := getObfs4CertString()
	if certError != nil {
		t.Fail()
		return
	}
	obfs4Transport := obfs4.OptimizerTransport{
		CertString: *certstring,
		IatMode:    0,
		Address:    "127.0.0.1:1234",
		Dialer:     dialer,
	}
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{obfs4Transport, &shadowTransport}
	strategy := NewRandomStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)

	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}

func TestOptimizerTransportRotateDial(t *testing.T) {
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
		Dialer:     dialer,
}
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{obfs4Transport, &shadowTransport}
	strategy := NewRotateStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)

	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}

func TestOptimizerTransportTrackDial(t *testing.T) {
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
		Dialer:     dialer,
}
	shadowTransport := shadow.NewTransport("1234", "CHACHA20-IETF-POLY1305", "127.0.0.1:1235")
	transports := []Transport{obfs4Transport, &shadowTransport}
	strategy := NewTrackStrategy(transports)
	optimizerTransport := NewOptimizerClient(transports, strategy)

	for i := 1; i <= 3; i++ {
		_, err := optimizerTransport.Dial()
		if err != nil {
			t.Fail()
		}
	}
}

func TestOptimizerTransportMinimizeDialDurationDial(t *testing.T) {
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

func getObfs4CertString() (*string, error) {
	usr, userError := user.Current()
	if userError != nil {
		return nil, userError
	}
	home := usr.HomeDir
	var fPath string
	if runtime.GOOS == "darwin" {
		fPath = path.Join(home, "shapeshifter-transports/stateDir/obfs4_bridgeline.txt")
	} else {
		fPath = path.Join(home, "gopath/src/github.com/OperatorFoundation/shapeshifter-transports/stateDir/obfs4_bridgeline.txt")
	}
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
func RunLocalObfs2Server() {
	//create a server
	config := obfs2.NewObfs2Transport()

	//call listen on the server
	serverListener := config.Listen("127.0.0.1:1237")
	if serverListener == nil {
		return
	}

	//Create Server connection and format it for concurrency
	go func() {
		//create server buffer
		serverBuffer := make([]byte, 4)

		//create serverConn
		for {
			serverConn, acceptErr := serverListener.Accept()
			if acceptErr != nil {
				return
			}

			//read on server side
			_, serverReadErr := serverConn.Read(serverBuffer)
			if serverReadErr != nil {
				return
			}

			//write data from serverConn for client to read
			_, serverWriteErr := serverConn.Write([]byte(data))
			if serverWriteErr != nil {
				return
			}
		}
	}()
}