/*
	MIT License

	Copyright (c) 2020 Operator Foundation

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

package polish

import (
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/crypto/chacha20poly1305"
	"io"
	"math/big"
	"net"
)

// Public key size in bytes.
// See "Marshal" in elliptic.go
// Marshal implements section 4.3.6 of ANSI X9.62
var silverPublicKeySize = 1+2*((elliptic.P256().Params().BitSize + 7) >> 3)

type SilverPolishClientConfig struct {
	ServerPublicKey []byte
	ChunkSize       int
}

type SilverPolishServerConfig struct {
	ServerPublicKey  []byte
	ServerPrivateKey []byte
	ChunkSize        int
}

func (config SilverPolishServerConfig) Construct() (Server, error) {
	return NewSilverServer(config)
}

func (config SilverPolishServerConfig) GetChunkSize() int {
	return config.ChunkSize
}

func (config SilverPolishClientConfig) Construct() (Connection, error) {
	return NewSilverClient(config)
}

func (config SilverPolishClientConfig) GetChunkSize() int {
return config.ChunkSize
}

type SilverPolishClient struct {
	serverPublicKey []byte
	chunkSize       int

	clientPublicKey []byte
	clientPrivateKey []byte

	sharedKey []byte
	polishCipher cipher.AEAD
}

type SilverPolishServer struct {
	serverPublicKey  []byte
	serverPrivateKey []byte
	chunkSize        int

	connections map[net.Conn]SilverPolishServerConnection
}

type SilverPolishServerConnection struct {
	serverPublicKey []byte
	serverPrivateKey []byte
	chunkSize int

	clientPublicKey []byte

	sharedKey []byte
	polishCipher cipher.AEAD
}

type CurvePoint struct {
	X *big.Int
	Y *big.Int
}

func NewSilverServerConfig() (*SilverPolishServerConfig, error) {
	curve := elliptic.P256()
	serverPrivateKey, serverX, serverY, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, errors.New("error generating server private key")
	}
	serverPublicKey := elliptic.Marshal(curve, serverX, serverY)

	tempClientPrivateKey, _, _, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, errors.New("error generating temporary client private key")
	}

	tempSharedKeyX, tempSharedKeyY := curve.ScalarMult(serverX, serverY, tempClientPrivateKey)
	tempSharedKeySeed := elliptic.Marshal(curve, tempSharedKeyX, tempSharedKeyY)

	// X963 KDF
	encryptionKey := X963KDF(tempSharedKeySeed, serverPublicKey)

	tempCipher, err := chacha20poly1305.New(encryptionKey)
	if err != nil {
		return nil, errors.New("error generating new config")
	}

	basePayloadSize := 1024
	payloadSizeRandomness, err := rand.Int(rand.Reader, big.NewInt(512))
	if err != nil {
		return nil, errors.New("error generating random number for ChunkSize")
	}

	payloadSize := basePayloadSize + int(payloadSizeRandomness.Int64())
	chunkSize := tempCipher.NonceSize() + tempCipher.Overhead() + payloadSize

	config := SilverPolishServerConfig{serverPublicKey, serverPrivateKey, chunkSize}
	return &config, nil
}

func NewSilverClientConfig(serverConfig *SilverPolishServerConfig) (*SilverPolishClientConfig, error) {
	config := SilverPolishClientConfig{serverConfig.ServerPublicKey, serverConfig.ChunkSize}
	return &config, nil
}

func NewSilverClient(config SilverPolishClientConfig) (Connection, error) {
	// Generate a new random private key
	curve := elliptic.P256()
	clientPrivateKey, clientX, clientY, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, errors.New("error generating client private key")
	}

	// Marshall uses section 4.3.6 of ANSI X9.62
	clientPublicKey := elliptic.Marshal(curve, clientX, clientY)

	// Derive the shared key from the client private key and server public key
	serverX, serverY := elliptic.Unmarshal(curve, config.ServerPublicKey)

	sharedKeyX, sharedKeyY := curve.ScalarMult(serverX, serverY, clientPrivateKey)
	sharedKeySeed := elliptic.Marshal(curve, sharedKeyX, sharedKeyY)

	encryptionKey := X963KDF(sharedKeySeed, clientPublicKey)

	polishCipher, err := chacha20poly1305.New(encryptionKey[:])
	if err != nil {
		return nil, errors.New("error initializing polish client")
	}
	polishClient := SilverPolishClient{config.ServerPublicKey, config.ChunkSize, clientPublicKey, clientPrivateKey, encryptionKey, polishCipher}
	return &polishClient, nil
}

func X963KDF(sharedKeySeed []byte, ephemeralPublicKey []byte) []byte {

	//FIXME: Is this a correct X963 KDF
	length := 32
	output := make([]byte, 0)
	outlen := 0
	counter := uint32(1)

	for outlen < length {
		h := sha256.New()
		h.Write(sharedKeySeed) // Key Material: ECDH Key

		counterBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(counterBuf, counter)
		h.Write(counterBuf)

		h.Write(ephemeralPublicKey) // Shared Info: Our public key

		output = h.Sum(output)
		outlen += h.Size()
		counter += 1
	}

	return output
}

func NewSilverServer(config SilverPolishServerConfig) (SilverPolishServer, error) {
	polishServer := SilverPolishServer{config.ServerPublicKey, config.ServerPrivateKey, config.ChunkSize, make(map[net.Conn]SilverPolishServerConnection)}
	return polishServer, nil
}

func (config SilverPolishServer) NewConnection(conn net.Conn) Connection {
	polishServerConnection := SilverPolishServerConnection{config.serverPublicKey, config.serverPrivateKey, config.chunkSize, nil, nil, nil}
	config.connections[conn] = polishServerConnection

	return &polishServerConnection
}

func (silver SilverPolishClient) Handshake(conn net.Conn) error {
	clientPublicKey := silver.clientPublicKey
	publicKeyBlock := make([]byte, silver.chunkSize)
	_, readError := rand.Read(publicKeyBlock)
	if readError != nil {
		return readError
	}
	copy(publicKeyBlock, clientPublicKey[:])
	_, writeError := conn.Write(publicKeyBlock)
	if writeError != nil {
		return writeError
	}

	return nil
}
func Polish(polishCipher cipher.AEAD, chunkSize int, input []byte) ([]byte, error) {
	output := make([]byte, 0)

	inputSize := len(input)
	nonceSize := polishCipher.NonceSize()
	overheadSize := polishCipher.Overhead()
	payloadSize := chunkSize - (nonceSize + overheadSize)
	maximumInputSize := payloadSize - 2
	payload := make([]byte, payloadSize)

	// Generate random nonce
	nonce := make([]byte, nonceSize)
	_, readError := rand.Read(nonce)
	if readError != nil {
		return nil, readError
	}

	// Payload size - 2 because the first two bytes indicate the length of the input data
	// These 2 bytes are not part of the data itself they describe its length
	// Payload = length(2 bytes) + input data + padding
	if inputSize <= maximumInputSize {

		inputLengthUInt16 := uint16(inputSize)
		// Convert it to two bytes and add it to the beginning of our payload
		binary.LittleEndian.PutUint16(payload, inputLengthUInt16)

		//Put input into out payload slice starting at the 3rd byte
		copy(payload[2:], input)

		// Encrypt the payload
		sealResult := polishCipher.Seal(output, nonce, payload, nil)
		result := append(nonce, sealResult...)

		return result, nil
	} else {
		//Input is larger than chunk size
		result := make([]byte, 0)

		for len(input) > 0 {

			// Make sure that we don't try to make a slice that is bigger than input
			chunk := make([]byte, 0)
			if len(input) < maximumInputSize {
				chunk = input[:]
				input = make([]byte, 0)
			} else {
				chunk = input[:maximumInputSize]
				input = input[maximumInputSize:]
			}

			// polish this chunk of input
			polished, polishError := Polish(polishCipher, chunkSize, chunk)
			if polishError != nil {
				return nil, polishError
			}

			result = append(result, polished...)
		}

		return result, nil
	}
}

func Unpolish(polishCipher cipher.AEAD, chunkSize int, input []byte) ([]byte, error) {
	inputSize := len(input)
	nonceSize := polishCipher.NonceSize()

	output := make([]byte, 0)
	nonce := input[:nonceSize]
	data := input[nonceSize:]

	if inputSize < chunkSize {
		return nil, errors.New("silver client - unable to unpolish data, received fewer bytes than chunk size")
	} else if inputSize == chunkSize {
		unpolished, openError := polishCipher.Open(output, nonce, data, nil)
		if openError != nil {
			return nil, openError
		}

		dataSize := int(binary.LittleEndian.Uint16(unpolished))
		data = unpolished[2:dataSize + 2]
		return data, nil
	} else {
		//More than one chunk
		result := make([]byte, 0)
		for len(input) > 0 {

			if len(input) < chunkSize {
				return nil, errors.New("received input to unpolish that is less than chunk size")
			}

			chunk := input[:chunkSize]
			input = input[chunkSize:]

			unpolished, unpolishError := Unpolish(polishCipher, chunkSize, chunk)
			if unpolishError != nil {
				return nil, unpolishError
			}

			result = append(result, unpolished...)
		}

		return result, nil
	}
}

func (silver SilverPolishClient) Polish(input []byte) ([]byte, error) {
	return Polish(silver.polishCipher, silver.chunkSize, input)
}

func (silver SilverPolishClient) Unpolish(input []byte) ([]byte, error) {
	return Unpolish(silver.polishCipher, silver.chunkSize, input)
}

func (silver SilverPolishClient) GetChunkSize() int {
	return silver.chunkSize
}

func (silver *SilverPolishServerConnection) Handshake(conn net.Conn) error {
	curve := elliptic.P256()

	clientPublicKeyBlock := make([]byte, silver.chunkSize)
	_, err := io.ReadFull(conn, clientPublicKeyBlock)
	if err != nil {
		log.Error(errors.New("handshake error initializing polish shared key"))
		log.Error(err)
		return err
	}

	clientPublicKey := make([]byte, silverPublicKeySize)
	copy(clientPublicKey[:], clientPublicKeyBlock[:silverPublicKeySize])

	// Marshall uses section 4.3.6 of ANSI X9.62
	clientX, clientY := elliptic.Unmarshal(curve, clientPublicKey[:])
	if clientX == nil || clientY == nil {
		unmarshallError := errors.New("silver server unmarshal error: received a nil response while decoding the client public key")
		log.Error(unmarshallError)
		return unmarshallError
	}

	sharedKeyX, sharedKeyY := curve.ScalarMult(clientX, clientY, silver.serverPrivateKey)
	sharedKeySeed := elliptic.Marshal(curve, sharedKeyX, sharedKeyY)

	sharedKey := X963KDF(sharedKeySeed, clientPublicKey)

	silver.polishCipher, err = chacha20poly1305.New(sharedKey)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (silver *SilverPolishServerConnection) Polish(input []byte) ([]byte, error) {
	return Polish(silver.polishCipher, silver.chunkSize, input)
}

func (silver *SilverPolishServerConnection) Unpolish(input []byte) ([]byte, error) {
	if silver.polishCipher != nil {
		return Unpolish(silver.polishCipher, silver.chunkSize, input)
	} else {
		nilCipherError := errors.New("unable to unpolish input, silver.polishCipher is nil")
		log.Error(nilCipherError)
		return nil, nilCipherError
	}
}

func (silver *SilverPolishServerConnection) GetChunkSize() int {
	return silver.chunkSize
}
