/*
 * Copyright (c) 2019, Operator Foundation
 *
 */

// Package Optimizer provides a PT 2.0 Go API wrapper around the connections used
package Optimizer

import (
	"github.com/OperatorFoundation/obfs4/common/drbg"
	"github.com/OperatorFoundation/obfs4/common/ntor"
	"github.com/OperatorFoundation/obfs4/common/replayfilter"
	pt "github.com/OperatorFoundation/shapeshifter-ipc"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/shadow"
	_ "github.com/OperatorFoundation/shapeshifter-transports/transports/shadow"
	"net"
)
type Transport interface {
	Dial() net.Conn
}

type optimizerTransport struct {
	transports []Transport
}

type ShadowTransport struct {
	password   string
	cipherName string
	address    string
}

type Obfs4Transport struct {
	serverFactory *Obfs4ServerFactory
	clientArgs 	  *Obfs4ClientArgs
}

type Obfs4ServerFactory struct {
	args *pt.Args

	nodeID       *ntor.NodeID
	identityKey  *ntor.Keypair
	lenSeed      *drbg.Seed
	iatSeed      *drbg.Seed
	iatMode      int
	replayFilter *replayfilter.ReplayFilter

	closeDelayBytes int
	closeDelay      int
}

type Obfs4ClientArgs struct {
	nodeID     *ntor.NodeID
	publicKey  *ntor.PublicKey
	sessionKey *ntor.Keypair
	iatMode    int
}

//I created this function in optimizers code because it doesn't exist in obfs4's code
func NewObfs4Transport(serverFactory *Obfs4ServerFactory, clientArgs *Obfs4ClientArgs) *Obfs4Transport {
	return &Obfs4Transport{serverFactory:serverFactory, clientArgs:clientArgs}
}

func NewOptimizerTransport(transports []Transport) *optimizerTransport {
	return &optimizerTransport{transports}
}

func NewOptimizerClient (transports []Transport) *optimizerTransport {
	return &optimizerTransport{transports}
}

func (opTransport *optimizerTransport) Dial() net.Conn {
	var transport Transport
	transport = opTransport.transports[0]

	conn := transport.Dial()
	if conn == nil {
		return nil
	}

	return conn
}

func (transport ShadowTransport) Dial() net.Conn {
	shadowTransport := shadow.NewShadowTransport(transport.password, transport.cipherName)
	conn := shadowTransport.Dial(transport.address)
	return conn
}

//Because obfs4 is structured way different, im having trouble making it compatible with shadows format.
//I tried removing the  "obfs4." and it helped recognize "NewObfs4Transport", but then it had a problem
//with the objects in the parenthesis, saying they were incompatible?
//Also, the dial function exists in obfs4, but the below function can't locate it, along with "address"
func (transport Obfs4Transport) Dial() net.Conn {
	Obfs4Transport := obfs4.NewObfs4Transport(transport.clientArgs, transport.serverFactory)
	conn := Obfs4Transport.Dial(transport.address)
	return conn
}