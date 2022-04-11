package StarBridge

import (
	"github.com/OperatorFoundation/monolith-go/monolith"
	pt "github.com/OperatorFoundation/shapeshifter-ipc/v2"
	replicant "github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2/polish"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2/toneburst"
	"golang.org/x/net/proxy"
	"math/rand"
	"net"
	"time"
)

type Transport struct {
	Config  ClientConfig
	Address string
	Dialer  proxy.Dialer
}

type ClientConfig struct {
	Address string `json:"serverAddress"`
}

type ServerConfig struct {
}

type starbridgeTransportListener struct {
	listener *net.TCPListener
	config   ServerConfig
}

func newStarBridgeTransportListener(listener *net.TCPListener, config ServerConfig) *starbridgeTransportListener {
	return &starbridgeTransportListener{listener: listener, config: config}
}

func (listener *starbridgeTransportListener) Addr() net.Addr {
	interfaces, _ := net.Interfaces()
	addrs, _ := interfaces[0].Addrs()
	return addrs[0]
}

// Accept waits for and returns the next connection to the listener.
func (listener *starbridgeTransportListener) Accept() (net.Conn, error) {
	conn, err := listener.listener.Accept()
	if err != nil {
		return nil, err
	}

	return NewServerConnection(conn)
}

// Close closes the transport listener.
// Any blocked Accept operations will be unblocked and return errors.
func (listener *starbridgeTransportListener) Close() error {
	return listener.listener.Close()
}

//Listen checks for a working connection
func (config ServerConfig) Listen(address string) net.Listener {
	addr, resolveErr := pt.ResolveAddr(address)
	if resolveErr != nil {
		return nil
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil
	}

	return newStarBridgeTransportListener(ln, config)
}

//Dial connects to the address on the named network
func (config ClientConfig) Dial(address string) (net.Conn, error) {
	conn, dialErr := net.Dial("tcp", address)
	if dialErr != nil {
		return nil, dialErr
	}

	transportConn, err := NewClientConnection(conn)
	if err != nil {
		if conn != nil {
			_ = conn.Close()
		}
		return nil, err
	}

	return transportConn, nil
}

// Dial creates outgoing transport connection
func (transport *Transport) Dial() (net.Conn, error) {
	conn, dialErr := transport.Dialer.Dial("tcp", transport.Address)
	if dialErr != nil {
		return nil, dialErr
	}

	dialConn := conn
	transportConn, err := NewClientConnection(conn)
	if err != nil {
		_ = dialConn.Close()
		return nil, err
	}

	return transportConn, nil
}

func NewClientConnection(conn net.Conn) (*replicant.Connection, error) {
	config := getClientConfig()
	return replicant.NewClientConnection(conn, config)
}

func NewServerConnection(conn net.Conn) (*replicant.Connection, error) {
	config := getServerConfig()
	return replicant.NewServerConnection(conn, config)
}

func NewReplicantClientConnectionState() (*replicant.ConnectionState, error) {
	config := getClientConfig()
	return replicant.NewReplicantClientConnectionState(config)
}

func NewReplicantServerConnectionState(polishServer polish.Server, conn net.Conn) (*replicant.ConnectionState, error) {
	config := getServerConfig()
	return replicant.NewReplicantServerConnectionState(config, polishServer, conn)
}

func getClientConfig() replicant.ClientConfig {
	rand.Seed(time.Now().UnixNano())

	clientDesc, serverDesc := createStarBridgeToneburstDescriptions()
	clientInstance := createStarBridgeToneburstClientInstance(clientDesc)

	// The client speaks second.
	monotoneClientConfig := toneburst.MonotoneConfig{
		AddSequences:    clientInstance,
		RemoveSequences: serverDesc,
		SpeakFirst:      false,
	}

	return replicant.ClientConfig{Toneburst: monotoneClientConfig}
}

func getServerConfig() replicant.ServerConfig {
	rand.Seed(time.Now().UnixNano())

	clientDesc, serverDesc := createStarBridgeToneburstDescriptions()
	serverInstance := createStarBridgeToneburstServerInstance(serverDesc)

	// The server speaks first.
	monotoneServerConfig := toneburst.MonotoneConfig{
		AddSequences:    serverInstance,
		RemoveSequences: clientDesc,
		SpeakFirst:      true,
	}

	return replicant.ServerConfig{Toneburst: monotoneServerConfig}
}

// Generate instances from descriptions.
// Instances are using for sending, they are fully specified.
// Note: this is where arguments would go if we add some
//TODO: split into two functions
func createStarBridgeToneburstClientInstance(clientDesc *monolith.Description) *monolith.Instance {
	clientInstance := monolith.Instance{
		Desc: *clientDesc,
		Args: monolith.NewEmptyArgs(),
	}

	return &clientInstance
}

func createStarBridgeToneburstServerInstance(serverDesc *monolith.Description) *monolith.Instance {
	serverInstance := monolith.Instance{
		Desc: *serverDesc,
		Args: monolith.NewEmptyArgs(),
	}

	return &serverInstance
}

// Generate the descriptions using the parts.
// Descriptions are for receiving, they are partially specified as some elements are not know ahead of time.
func createStarBridgeToneburstDescriptions() (*monolith.Description, *monolith.Description) {
	clientParts, serverParts := createStarBridgeToneburstParts()

	clientDesc := monolith.Description{clientParts}
	serverDesc := monolith.Description{serverParts}

	return &clientDesc, &serverDesc
}

// Generate the parts.
// Each part results in one network send from one side of the connection.
// On the other side of the connection, the data is received, but it may come in as multiple packets.
// The client and server take turns sending, one part at a time.
// FIXME: Currently, this config just has one part in each direction, but this probably needs to change.
// Note: each part is an individual packet of the client-server convo
func createStarBridgeToneburstParts() ([]monolith.Monolith, []monolith.Monolith) {
	part1 := createStarBridgeToneburstServerBytesPart1()
	part3 := createStarBridgeToneburstServerBytesPart3()
	part5 := createStarBridgeToneburstServerBytesPart5()

	serverParts := []monolith.Monolith{
		part1,
		part3,
		part5,
	}

	part2 := createStarBridgeToneburstClientBytesPart2()
	part4 := createStarBridgeToneburstClientBytesPart4()

	clientParts := []monolith.Monolith{
		part2,
		part4,
	}

	return clientParts, serverParts
}

// Generate part 1, the first thing the server says to the client. The server speaks first.
// Example: S: 220 mail.imc.org SMTP service ready
// This part has 4 subsections:
// a. 220 - fixed
// b. mail.imc.org - variable
// c. SMTP - fixed
// d. service ready - ???
func createStarBridgeToneburstServerBytesPart1() monolith.StringsPart {

	serverPart := monolith.StringsPart{
		Items: []monolith.StringType{
			monolith.FixedStringType{String: "220 "},
			// FIXME: needs to be a variable string type
			monolith.FixedStringType{String: "mail.imc.org "},
			monolith.FixedStringType{String: "SMTP service ready\r\n"},
		},
	}

	return serverPart
}

// Generate part 2, the first thing that the client says to the server. The server speaks first.
// Example: C: EHLO mail.example.com
// This part has 4 subsections:
// a. EHLO - fixed
// b. mail.example.com - variable
func createStarBridgeToneburstClientBytesPart2() monolith.StringsPart {
	clientPart := monolith.StringsPart{
		Items: []monolith.StringType{
			monolith.FixedStringType{String: "EHLO "},
			// FIXME: eventually needs to be variable string type
			monolith.FixedStringType{String: "mail.imc.org\r\n"},
		},
	}

	return clientPart
}

// We need at least three more parts

/* Part 3
   S: 250-mail.imc.org offers a warm hug of welcome
   S: 250-8BITMIME
   S: 250-STARTTLS
   S: 250 DSN

   This section has 4 subsections:
   a. "250-" - fixed
   b. mail.imc.org - variable, but the same as part 1
   c. "offers a warm hug of welcome" - variable
   c. 250-8BITMIME...etc. - variable. There are several options for lines that can appear here.
*/
func createStarBridgeToneburstServerBytesPart3() monolith.StringsPart {
	serverPart := monolith.StringsPart{
		Items: []monolith.StringType{
			monolith.FixedStringType{String: "250-"},
			// FIXME: needs to eventually be variable string type
			monolith.FixedStringType{String: "mail.imc.org"},
			monolith.FixedStringType{String: " offers a warm hug of welcome\r\n"},
			monolith.FixedStringType{String: "250-"},
			// FIXME: needs to eventually be variable string type
			monolith.FixedStringType{String: "8BITMIME\r\n"},
			monolith.FixedStringType{String: "250-"},
			// FIXME: needs to eventually be variable string type
			monolith.FixedStringType{String: "STARTTLS\r\n"},
			monolith.FixedStringType{String: "250 "},
			// FIXME: needs to eventually be variable string type
			monolith.FixedStringType{String: "DSN\r\n"},
		},
	}

	return serverPart
}

/* Part 4
   C: STARTTLS

   This part is only one section.
   a.STARTTLS - fixed
*/
func createStarBridgeToneburstClientBytesPart4() monolith.StringsPart {
	clientPart := monolith.StringsPart{
		Items: []monolith.StringType{
			monolith.FixedStringType{String: "STARTTLS\r\n"},
		},
	}

	return clientPart
}

/* Part 5
   S: 220 Go ahead

   This part has two subsections.
   a. 220 - FIXED
   b. "Go ahead" - variable
*/
func createStarBridgeToneburstServerBytesPart5() monolith.StringsPart {
	serverPart := monolith.StringsPart{
		Items: []monolith.StringType{
			monolith.FixedStringType{String: "220 "},
			// FIXME: eventually needs to be variable string type
			monolith.FixedStringType{String: "Go ahead\r\n"},
		},
	}

	return serverPart
}
