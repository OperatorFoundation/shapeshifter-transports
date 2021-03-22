/*
* Copyright (c) 2014, Yawning Angel <yawning at torproject dot org>
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions are met:
*
*  * Redistributions of source code must retain the above copyright notice,
*    this list of conditions and the following disclaimer.
*
*  * Redistributions in binary form must reproduce the above copyright notice,
*    this list of conditions and the following disclaimer in the documentation
*    and/or other materials provided with the distribution.
*
* THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
* AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
* IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
* ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
* LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
* CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
* SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
* INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
* CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
* ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
* POSSIBILITY OF SUCH DAMAGE.
*/

// Package obfs4 provides an implementation of the Tor Project's obfs4
// obfuscation protocol.
package obfs4

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/OperatorFoundation/obfs4/common/drbg"
	"github.com/OperatorFoundation/obfs4/common/log"
	"github.com/OperatorFoundation/obfs4/common/ntor"
	"github.com/OperatorFoundation/obfs4/common/probdist"
	"github.com/OperatorFoundation/obfs4/common/replayfilter"
	"github.com/OperatorFoundation/shapeshifter-ipc/v2"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4/v2/framing"
	"golang.org/x/net/proxy"
	"math/rand"
	"net"
	"strconv"
	"syscall"
	"time"
)

const (
	nodeIDArg     = "node-id"
	privateKeyArg = "private-key"
	seedArg       = "drbg-seed"
	iatArg        = "iat-mode"
	certArg       = "cert"

	biasCmdArg = "obfs4-distBias"

	seedLength             = drbg.SeedLength
	headerLength           = framing.FrameOverhead + packetOverhead
	clientHandshakeTimeout = time.Duration(60) * time.Second
	serverHandshakeTimeout = time.Duration(30) * time.Second
	replayTTL              = time.Duration(3) * time.Hour

	maxIATDelay        = 100
	maxCloseDelayBytes = maxHandshakeLength
	maxCloseDelay      = 60
)

const (
	iatNone = iota
	iatEnabled
	iatParanoid
)

// biasedDist controls if the probability table will be ScrambleSuit style or
// uniformly distributed.
var biasedDist bool

// Transport that uses the obfs4 protocol to shapeshift the application network traffic
type Transport struct {
	dialer proxy.Dialer

	serverFactory *ServerFactory
	clientArgs    *ClientArgs
}

//ServerFactory contains arguments for server side
type ServerFactory struct {
	args map[string]string

	nodeID       *ntor.NodeID
	identityKey  *ntor.Keypair
	lenSeed      *drbg.Seed
	iatSeed      *drbg.Seed
	iatMode      int
	replayFilter *replayfilter.ReplayFilter

	closeDelayBytes int
	closeDelay      int
}

//ClientArgs contains arguments for client side
type ClientArgs struct {
	nodeID     *ntor.NodeID
	publicKey  *ntor.PublicKey
	sessionKey *ntor.Keypair
	iatMode    int
}

//NewObfs4Server initializes the obfs4 server side
func NewObfs4Server(stateDir string) (*Transport, error) {
	args := make(map[string]string)
	st, err := serverStateFromArgs(stateDir, args)
	if err != nil {
		return nil, err
	}

	var iatSeed *drbg.Seed
	if st.iatMode != iatNone {
		iatSeedSrc := sha256.Sum256(st.drbgSeed.Bytes()[:])
		var err error
		iatSeed, err = drbg.SeedFromBytes(iatSeedSrc[:])
		if err != nil {
			return nil, err
		}
	}

	// Store the arguments that should appear in our descriptor for the clients.
	ptArgs := make(map[string]string)
	ptArgs[certArg] = st.cert.String()
	log.Infof("certstring %s", certArg)
	ptArgs[iatArg] = strconv.Itoa(st.iatMode)

	// Initialize the replay filter.
	filter, err := replayfilter.New(replayTTL)
	if err != nil {
		return nil, err
	}

	// Initialize the close thresholds for failed connections.
	hashDrbg, err := drbg.NewHashDrbg(st.drbgSeed)
	if err != nil {
		return nil, err
	}
	rng := rand.New(hashDrbg)

	sf := &ServerFactory{ptArgs, st.nodeID, st.identityKey, st.drbgSeed, iatSeed, st.iatMode, filter, rng.Intn(maxCloseDelayBytes), rng.Intn(maxCloseDelay)}

	return &Transport{dialer: nil, serverFactory: sf, clientArgs: nil}, nil
}

//NewObfs4Client initializes the obfs4 client side
func NewObfs4Client(certString string, iatMode int, dialer proxy.Dialer) (*Transport, error) {
	var nodeID *ntor.NodeID
	var publicKey *ntor.PublicKey

	// The "new" (version >= 0.0.3) bridge lines use a unified "cert" argument
	// for the Node ID and Public Key.
	cert, err := serverCertFromString(certString)
	if err != nil {
		return nil, err
	}
	nodeID, publicKey = cert.unpack()

	// Generate the session key pair before connection to hide the Elligator2
	// rejection sampling from network observers.
	sessionKey, err := ntor.NewKeypair(true)
	if err != nil {
		return nil, err
	}

	if dialer == nil {
		return &Transport{dialer: proxy.Direct, serverFactory: nil, clientArgs: &ClientArgs{nodeID, publicKey, sessionKey, iatMode}}, nil
	}
		return &Transport{dialer: dialer, serverFactory: nil, clientArgs: &ClientArgs{nodeID, publicKey, sessionKey, iatMode}}, nil

}

// Dial creates outgoing transport connection
func (transport *Transport) Dial(address string) (net.Conn, error) {
	dialFn := transport.dialer.Dial
	conn, dialErr := dialFn("tcp", address)
	if dialErr != nil {
		return nil, dialErr
	}

	dialConn := conn
	transportConn, err := newObfs4ClientConn(conn, transport.clientArgs)
	if err != nil {
		closeErr := dialConn.Close()
		if closeErr != nil {
			log.Errorf("could not close")
		}
		return nil, err
	}

	return transportConn, nil
}


//OptimizerTransport contains parameters to be used in Optimizer
type OptimizerTransport struct {
	CertString string
	IatMode    int
	Address    string
	Dialer     proxy.Dialer
}

//Config contains arguments formatted for a json file
type Config struct {
	CertString string `json:"cert"`
	IatMode    string `json:"iat-mode"`
}

// Dial creates outgoing transport connection
func (transport OptimizerTransport) Dial() (net.Conn, error) {
	Obfs4Transport, err := NewObfs4Client(transport.CertString, transport.IatMode, transport.Dialer)
	if err != nil {
		return nil, err
	}
	conn, err := Obfs4Transport.Dial(transport.Address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Listen creates listener for incoming transport connection
func (transport *Transport) Listen(address string) net.Listener {
	addr, resolveErr := pt.ResolveAddr(address)
	if resolveErr != nil {
		fmt.Println(resolveErr.Error())
		return nil
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return newObfs4TransportListener(transport.serverFactory, ln)
}

// Close closes the transport listener.
func (transport *Transport) Close() error {
	return nil
}

// End methods that implement the base.Transport interface

// Listener that accepts connections using the obfs4 transport to communicate
type obfs4TransportListener struct {
	serverFactory *ServerFactory

	listener *net.TCPListener
}

// Private initializer for the obfs4 listener.
// You get a new listener instance by calling the Listen method on the Transport.
func newObfs4TransportListener(sf *ServerFactory, listener *net.TCPListener) *obfs4TransportListener {
	return &obfs4TransportListener{serverFactory: sf, listener: listener}
}

// Methods that implement the net.Listener interface

// NetworkListener listens for underlying network connection
func (listener *obfs4TransportListener) NetworkListener() net.Listener {
	return listener.listener
}

// Accept waits for and returns the next connection to the listener.
func (listener *obfs4TransportListener) Accept() (net.Conn, error) {
	conn, err := listener.listener.Accept()
	if err != nil {
		return nil, err
	}

	return newObfs4ServerConn(conn, listener.serverFactory)
}

func (listener *obfs4TransportListener) Addr() net.Addr {
	interfaces, _ := net.Interfaces()
	addrs, _ := interfaces[0].Addrs()
	return addrs[0]
}

// Close closes the transport listener.
// Any blocked Accept operations will be unblocked and return errors.
func (listener *obfs4TransportListener) Close() error {
	return listener.listener.Close()
}

// End methods that implement the net.Listener interface

// Implementation of net.Conn, which also requires implementing net.Conn
type obfs4Conn struct {
	net.Conn

	isServer bool

	lenDist *probdist.WeightedDist
	iatDist *probdist.WeightedDist
	iatMode int

	receiveBuffer        *bytes.Buffer
	receiveDecodedBuffer *bytes.Buffer
	readBuffer           []byte

	encoder *framing.Encoder
	decoder *framing.Decoder
}

// Private initializer methods

func newObfs4ClientConn(conn net.Conn, args *ClientArgs) (c *obfs4Conn, err error) {
	// Generate the initial protocol polymorphism distribution(s).
	var seed *drbg.Seed
	if seed, err = drbg.NewSeed(); err != nil {
		return
	}
	lenDist := probdist.New(seed, 0, framing.MaximumSegmentLength, biasedDist)
	var iatDist *probdist.WeightedDist
	if args.iatMode != iatNone {
		var iatSeed *drbg.Seed
		iatSeedSrc := sha256.Sum256(seed.Bytes()[:])
		if iatSeed, err = drbg.SeedFromBytes(iatSeedSrc[:]); err != nil {
			return
		}
		iatDist = probdist.New(iatSeed, 0, maxIATDelay, biasedDist)
	}

	// Allocate the client structure.
	c = &obfs4Conn{conn, false, lenDist, iatDist, args.iatMode, bytes.NewBuffer(nil), bytes.NewBuffer(nil), make([]byte, consumeReadSize), nil, nil}

	// Start the handshake timeout.
	deadline := time.Now().Add(clientHandshakeTimeout)
	if err = conn.SetDeadline(deadline); err != nil {
		return nil, err
	}

	if err = c.clientHandshake(args.nodeID, args.publicKey, args.sessionKey); err != nil {
		return nil, err
	}

	// Stop the handshake timeout.
	if err = conn.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}

	return
}

func newObfs4ServerConn(conn net.Conn, sf *ServerFactory) (*obfs4Conn, error) {
	// Not much point in having a separate newObfs4ServerConn routine when
	// wrapping requires using values from the factory instance.

	// Generate the session keypair *before* consuming data from the peer, to
	// attempt to mask the rejection sampling due to use of Elligator2.  This
	// might be futile, but the timing differential isn't very large on modern
	// hardware, and there are far easier statistical attacks that can be
	// mounted as a distinguisher.
	sessionKey, err := ntor.NewKeypair(true)
	if err != nil {
		return nil, err
	}

	lenDist := probdist.New(sf.lenSeed, 0, framing.MaximumSegmentLength, biasedDist)
	var iatDist *probdist.WeightedDist
	if sf.iatSeed != nil {
		iatDist = probdist.New(sf.iatSeed, 0, maxIATDelay, biasedDist)
	}

	c := &obfs4Conn{conn, true, lenDist, iatDist, sf.iatMode, bytes.NewBuffer(nil), bytes.NewBuffer(nil), make([]byte, consumeReadSize), nil, nil}

	startTime := time.Now()

	if err = c.serverHandshake(sf, sessionKey); err != nil {
		c.closeAfterDelay(sf, startTime)
		return nil, err
	}

	return c, nil
}

// End initializer methods

// Methods that implement the net.Conn interface
func (transportConn *obfs4Conn) NetworkConn() net.Conn {
	return transportConn.Conn
}

// End methods that implement the net.Conn interface

// Methods implementing net.Conn
func (transportConn *obfs4Conn) Read(b []byte) (n int, err error) {
	// If there is no payload from the previous Read() calls, consume data off
	// the network.  Not all data received is guaranteed to be usable payload,
	// so do this in a loop till data is present or an error occurs.
	for transportConn.receiveDecodedBuffer.Len() == 0 {
		err = transportConn.readPackets()
		if err == framing.ErrAgain {
			// Don't propagate this back up the call stack if we happen to break
			// out of the loop.
			err = nil
			continue
		} else if err != nil {
			break
		}
	}

	// Even if err is set, attempt to do the read anyway so that all decoded
	// data gets relayed before the connection is torn down.
	if transportConn.receiveDecodedBuffer.Len() > 0 {
		var berr error
		n, berr = transportConn.receiveDecodedBuffer.Read(b)
		if err == nil {
			// Only propagate berr if there are not more important (fatal)
			// errors from the network/crypto/packet processing.
			err = berr
		}
	}

	return
}

func (transportConn *obfs4Conn) Write(b []byte) (n int, err error) {
	chopBuf := bytes.NewBuffer(b)
	var payload [maxPacketPayloadLength]byte
	var frameBuf bytes.Buffer

	// Chop the pending data into payload frames.
	for chopBuf.Len() > 0 {
		// Send maximum sized frames.
		rdLen := 0
		rdLen, err = chopBuf.Read(payload[:])
		if err != nil {
			return 0, err
		} else if rdLen == 0 {
			panic(fmt.Sprintf("BUG: Write(), chopping length was 0"))
		}
		n += rdLen

		err = transportConn.makePacket(&frameBuf, packetTypePayload, payload[:rdLen], 0)
		if err != nil {
			return 0, err
		}
	}

	if transportConn.iatMode != iatParanoid {
		// For non-paranoid IAT, pad once per burst.  Paranoid IAT handles
		// things differently.
		if err = transportConn.padBurst(&frameBuf, transportConn.lenDist.Sample()); err != nil {
			return 0, err
		}
	}

	// Write the pending data onto the network.  Partial writes are fatal,
	// because the frame encoder state is advanced, and the code doesn't keep
	// frameBuf around.  In theory, write timeouts and whatnot could be
	// supported if this wasn't the case, but that complicates the code.
	if transportConn.iatMode != iatNone {
		var iatFrame [framing.MaximumSegmentLength]byte
		for frameBuf.Len() > 0 {
			iatWrLen := 0

			switch transportConn.iatMode {
			case iatEnabled:
				// Standard (ScrambleSuit-style) IAT obfuscation optimizes for
				// bulk transport and will write ~MTU sized frames when
				// possible.
				iatWrLen, err = frameBuf.Read(iatFrame[:])

			case iatParanoid:
				// Paranoid IAT obfuscation throws performance out of the
				// window and will sample the length distribution every time a
				// write is scheduled.
				targetLen := transportConn.lenDist.Sample()
				if frameBuf.Len() < targetLen {
					// There's not enough data buffered for the target write,
					// so padding must be inserted.
					if err = transportConn.padBurst(&frameBuf, targetLen); err != nil {
						return 0, err
					}
					if frameBuf.Len() != targetLen {
						// Ugh, padding came out to a value that required more
						// than one frame, this is relatively unlikely so just
						// re-sample since there's enough data to ensure that
						// the next sample will be written.
						continue
					}
				}
				iatWrLen, err = frameBuf.Read(iatFrame[:targetLen])
			}
			if err != nil {
				return 0, err
			} else if iatWrLen == 0 {
				panic(fmt.Sprintf("BUG: Write(), iat length was 0"))
			}

			// Calculate the delay.  The delay resolution is 100 usec, leading
			// to a maximum delay of 10 msec.
			iatDelta := time.Duration(transportConn.iatDist.Sample() * 100)

			// Write then sleep.
			_, err = transportConn.Conn.Write(iatFrame[:iatWrLen])

			log.Debugf("Obfs4 Write called")
			log.Debugf(string(iatWrLen))
			if err != nil {
				return 0, err
			}
			time.Sleep(iatDelta * time.Microsecond)
		}
	} else {
		_, err = transportConn.Conn.Write(frameBuf.Bytes())
		log.Debugf("Obfs4 Write called")
		log.Debugf(string(len(frameBuf.Bytes())))
	}

	return
}

func (transportConn *obfs4Conn) Close() error {
	return transportConn.Conn.Close()
}

func (transportConn *obfs4Conn) LocalAddr() net.Addr {
	return transportConn.Conn.LocalAddr()
}

func (transportConn *obfs4Conn) RemoteAddr() net.Addr {
	return transportConn.Conn.RemoteAddr()
}

func (transportConn *obfs4Conn) SetDeadline(time.Time) error {
	return syscall.ENOTSUP
}

func (transportConn *obfs4Conn) SetReadDeadline(t time.Time) error {
	return transportConn.Conn.SetReadDeadline(t)
}

func (transportConn *obfs4Conn) SetWriteDeadline(time.Time) error {
	return syscall.ENOTSUP
}

// End of methods implementing net.Conn

// Private methods implementing the obfs4 protocol

func (transportConn *obfs4Conn) clientHandshake(nodeID *ntor.NodeID, peerIdentityKey *ntor.PublicKey, sessionKey *ntor.Keypair) error {
	if transportConn.isServer {
		return fmt.Errorf("clientHandshake called on server connection")
	}

	// Generate and send the client handshake.
	hs := newClientHandshake(nodeID, peerIdentityKey, sessionKey)
	blob, err := hs.generateHandshake()
	if err != nil {
		return err
	}
	if _, err = transportConn.Conn.Write(blob); err != nil {
		return err
	}

	// Consume the server handshake.
	var hsBuf [maxHandshakeLength]byte
	for {
		n, err := transportConn.Conn.Read(hsBuf[:])
		if err != nil {
			// The Read() could have returned data and an error, but there is
			// no point in continuing on an EOF or whatever.
			return err
		}
		transportConn.receiveBuffer.Write(hsBuf[:n])

		n, seed, err := hs.parseServerHandshake(transportConn.receiveBuffer.Bytes())
		if err == ErrMarkNotFoundYet {
			continue
		} else if err != nil {
			return err
		}
		_ = transportConn.receiveBuffer.Next(n)

		// Use the derived key material to initialize the link crypto.
		okm := ntor.Kdf(seed, framing.KeyLength*2)
		transportConn.encoder = framing.NewEncoder(okm[:framing.KeyLength])
		transportConn.decoder = framing.NewDecoder(okm[framing.KeyLength:])

		return nil
	}
}

func (transportConn *obfs4Conn) serverHandshake(sf *ServerFactory, sessionKey *ntor.Keypair) error {
	if !transportConn.isServer {
		return fmt.Errorf("serverHandshake called on client connection")
	}

	// Generate the server handshake, and arm the base timeout.
	hs := newServerHandshake(sf.nodeID, sf.identityKey, sessionKey)
	if err := transportConn.Conn.SetDeadline(time.Now().Add(serverHandshakeTimeout)); err != nil {
		return err
	}

	// Consume the client handshake.
	var hsBuf [maxHandshakeLength]byte
	for {
		n, err := transportConn.Conn.Read(hsBuf[:])
		if err != nil {
			// The Read() could have returned data and an error, but there is
			// no point in continuing on an EOF or whatever.
			return err
		}
		transportConn.receiveBuffer.Write(hsBuf[:n])

		seed, err := hs.parseClientHandshake(sf.replayFilter, transportConn.receiveBuffer.Bytes())
		if err == ErrMarkNotFoundYet {
			continue
		} else if err != nil {
			return err
		}
		transportConn.receiveBuffer.Reset()

		if err := transportConn.Conn.SetDeadline(time.Time{}); err != nil {
			return nil
		}

		// Use the derived key material to initialize the link crypto.
		okm := ntor.Kdf(seed, framing.KeyLength*2)
		transportConn.encoder = framing.NewEncoder(okm[framing.KeyLength:])
		transportConn.decoder = framing.NewDecoder(okm[:framing.KeyLength])

		break
	}

	// Since the current and only implementation always sends a PRNG seed for
	// the length obfuscation, this makes the amount of data received from the
	// server inconsistent with the length sent from the client.
	//
	// Re-balance this by tweaking the client minimum padding/server maximum
	// padding, and sending the PRNG seed un-padded (As in, treat the PRNG seed
	// as part of the server response).  See inlineSeedFrameLength in
	// handshake_ntor.go.

	// Generate/send the response.
	blob, err := hs.generateHandshake()
	if err != nil {
		return err
	}
	var frameBuf bytes.Buffer
	if _, err = frameBuf.Write(blob); err != nil {
		return err
	}

	// Send the PRNG seed as the first packet.
	if err := transportConn.makePacket(&frameBuf, packetTypePrngSeed, sf.lenSeed.Bytes()[:], 0); err != nil {
		return err
	}
	if _, err = transportConn.Conn.Write(frameBuf.Bytes()); err != nil {
		return err
	}

	return nil
}

func (transportConn *obfs4Conn) closeAfterDelay(sf *ServerFactory, startTime time.Time) {
	// I-it's not like I w-wanna handshake with you or anything.  B-b-baka!
	//defer transportConn.Conn.Close()

	delay := time.Duration(sf.closeDelay)*time.Second + serverHandshakeTimeout
	deadline := startTime.Add(delay)
	if time.Now().After(deadline) {
		return
	}

	if err := transportConn.Conn.SetReadDeadline(deadline); err != nil {
		return
	}

	// Consume and discard data on this connection until either the specified
	// interval passes or a certain size has been reached.
	discarded := 0
	var buf [framing.MaximumSegmentLength]byte
	for discarded < sf.closeDelayBytes {
		n, err := transportConn.Conn.Read(buf[:])
		if err != nil {
			return
		}
		discarded += n
	}
}

func (transportConn *obfs4Conn) padBurst(burst *bytes.Buffer, toPadTo int) (err error) {
	tailLen := burst.Len() % framing.MaximumSegmentLength

	padLen := 0
	if toPadTo >= tailLen {
		padLen = toPadTo - tailLen
	} else {
		padLen = (framing.MaximumSegmentLength - tailLen) + toPadTo
	}

	if padLen > headerLength {
		err = transportConn.makePacket(burst, packetTypePayload, []byte{},
			uint16(padLen-headerLength))
		if err != nil {
			return
		}
	} else if padLen > 0 {
		err = transportConn.makePacket(burst, packetTypePayload, []byte{},
			maxPacketPayloadLength)
		if err != nil {
			return
		}
		err = transportConn.makePacket(burst, packetTypePayload, []byte{},
			uint16(padLen))
		if err != nil {
			return
		}
	}

	return
}

func init() {
	flag.BoolVar(&biasedDist, biasCmdArg, false, "Enable obfs4 using ScrambleSuit style table generation")
}

// End private methods implementing the obfs4 protocol

// Force type checks to make sure that instances conform to interfaces
var _ net.Listener = (*obfs4TransportListener)(nil)
var _ net.Conn = (*obfs4Conn)(nil)
