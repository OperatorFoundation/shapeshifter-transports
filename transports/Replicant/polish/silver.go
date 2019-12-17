package polish

import (
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
	"io"
	"math/big"
	"net"
)

type SilverPolishConfig struct {
	ClientOrServer bool

	ClientConfig *SilverPolishClientConfig
	ServerConfig *SilverPolishServerConfig
}

type SilverPolishClientConfig struct {
	ServerPublicKey []byte
	ChunkSize       int
}

type SilverPolishServerConfig struct {
	ServerPublicKey  []byte
	ServerPrivateKey []byte
	ChunkSize        int
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

func NewSilverServerConfig() *SilverPolishServerConfig {
	curve := elliptic.P256()
	serverPrivateKey, serverX, serverY, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println("Error generating server private key")
		return nil
	}
	serverPublicKey := elliptic.Marshal(curve, serverX, serverY)

	tempClientPrivateKey, _, _, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println("Error generating temporary client private key")
		return nil
	}

	tempSharedKeyX, tempSharedKeyY := curve.ScalarMult(serverX, serverY, tempClientPrivateKey)
	tempSharedKeySeed := elliptic.Marshal(curve, tempSharedKeyX, tempSharedKeyY)

	// X963 KDF
	// FIXME
	encryptionKey := X963KDF(tempSharedKeySeed, serverPublicKey)

	//hasher := sha256.New
	//kdf := hkdf.New(hasher, tempSharedKeySeed, nil, nil)
	//tempSharedKey := make([]byte, chacha20poly1305.KeySize)
	//_, err = kdf.Read(tempSharedKey)
	//if err != nil {
	//	fmt.Println("Error hashing key seed for temporary shared key while making new config")
	//	return nil
	//}

	tempCipher, err := chacha20poly1305.New(encryptionKey)
	if err != nil {
		fmt.Println("Error generating new config")
		fmt.Println(err)
		return nil
	}

	basePayloadSize := 1024
	payloadSizeRandomness, err := rand.Int(rand.Reader, big.NewInt(512))
	if err != nil {
		fmt.Println("Error generating random number for ChunkSize")
		return nil
	}

	payloadSize := basePayloadSize + int(payloadSizeRandomness.Int64())
	chunkSize := tempCipher.NonceSize() + tempCipher.Overhead() + payloadSize

	config := SilverPolishServerConfig{serverPublicKey, serverPrivateKey, chunkSize}
	return &config
}

func NewSilverClientConfig(serverConfig *SilverPolishServerConfig) *SilverPolishClientConfig {
	config := SilverPolishClientConfig{serverConfig.ServerPublicKey, serverConfig.ChunkSize}
	return &config
}

func NewSilverClient(config SilverPolishClientConfig) *SilverPolishClient {
	// Generate a new random private key
	curve := elliptic.P256()
	clientPrivateKey, clientX, clientY, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println("Error generating client private key")
		return nil
	}
	clientPublicKey := elliptic.Marshal(curve, clientX, clientY)

	// Derive the shared key from the client private key and server public key
	serverX, serverY := elliptic.Unmarshal(curve, config.ServerPublicKey)

	sharedKeyX, sharedKeyY := curve.ScalarMult(serverX, serverY, clientPrivateKey)
	sharedKeySeed := elliptic.Marshal(curve, sharedKeyX, sharedKeyY)

	encryptionKey := X963KDF(sharedKeySeed, clientPublicKey)

	// FIXME: Is this a correct replacement?
	//hasher := sha256.New
	//kdf := hkdf.New(hasher, sharedKeySeed, nil, nil)
	//sharedKey := make([]byte, chacha20poly1305.KeySize)
	//kdf.Read(sharedKey)

	polishCipher, err := chacha20poly1305.New(encryptionKey[:])
	if err != nil {
		fmt.Println("Error initializing polish client")
		return nil
	}
	polishClient := SilverPolishClient{config.ServerPublicKey, config.ChunkSize, clientPublicKey, clientPrivateKey, encryptionKey, polishCipher}
	return &polishClient
}

func X963KDF(sharedKeySeed []byte, ephemeralPublicKey []byte) []byte {

	// X963 KDF
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

	// Key
	encryptionKey := output[0:16]
	iv := output[16:]

	fmt.Println("Created an encryption key and iv for the Silver client:")
	fmt.Println(hex.EncodeToString(encryptionKey))
	fmt.Println(hex.EncodeToString(iv))

	return output
	//
	//// AES
	//
	//block, _ := aes.NewCipher(encryptionKey)
	//aesgcm, _ := cipher.NewGCMWithNonceSize(block, 16)
	//
	//ct := aesgcm.Seal(nil, iv, []byte("Hello World"), nil)
	//
	//fmt.Println(hex.EncodeToString(ephemeralPublicKey) + hex.EncodeToString(ct))
}

func NewSilverServer(config SilverPolishServerConfig) *SilverPolishServer {
	polishServer := SilverPolishServer{config.ServerPublicKey, config.ServerPrivateKey, config.ChunkSize, make(map[net.Conn]SilverPolishServerConnection)}
	return &polishServer
}

func (config SilverPolishServer) NewConnection(conn net.Conn) PolishConnection {
	polishServerConnection := SilverPolishServerConnection{config.serverPublicKey, config.serverPrivateKey, config.chunkSize, nil, nil, nil}
	config.connections[conn] = polishServerConnection

	return polishServerConnection
}

func (silver SilverPolishClient) Handshake(conn net.Conn) error {
	clientPublicKey := silver.clientPublicKey
	publicKeyBlock := make([]byte, silver.chunkSize)
	rand.Read(publicKeyBlock)
	copy(publicKeyBlock, clientPublicKey[:])

	return nil
}

func (silver SilverPolishClient) Polish(input []byte) []byte {
	var output []byte

	// Generate random nonce
	nonce := make([]byte, silver.polishCipher.NonceSize())
	rand.Read(nonce)

	silver.polishCipher.Seal(output, nonce, input, nil)

	result := append(nonce, output...)

	return result
}

func (silver SilverPolishClient) Unpolish(input []byte) []byte {
	var output []byte

	nonceSize := silver.polishCipher.NonceSize()
	nonce := input[:nonceSize]
	data := input[nonceSize:]

	silver.polishCipher.Open(output, nonce, data, nil)

	return output
}

func (silver SilverPolishServerConnection) Handshake(conn net.Conn) error {
	curve := elliptic.P256()

	clientPublicKeyBlock := make([]byte, silver.chunkSize)
	_, err := io.ReadFull(conn, clientPublicKeyBlock)
	if err != nil {
		fmt.Println("Error initializing polish shared key")
		return err
	}

	clientPublicKey := &[chacha20poly1305.KeySize]byte{}
	copy(clientPublicKey[:], clientPublicKeyBlock[:chacha20poly1305.KeySize])

	clientX, clientY := elliptic.Unmarshal(curve, clientPublicKey[:])

	sharedKeyX, sharedKeyY := curve.ScalarMult(clientX, clientY, silver.serverPrivateKey)
	sharedKeySeed := elliptic.Marshal(curve, sharedKeyX, sharedKeyY)

	hasher := sha256.New
	kdf := hkdf.New(hasher, sharedKeySeed, nil, nil)
	sharedKey := make([]byte, chacha20poly1305.KeySize)
	kdf.Read(sharedKey)

	silver.polishCipher, err = chacha20poly1305.New(sharedKey)
	if err != nil {
		fmt.Println("Error initializing polish client")
		return nil
	}

	return nil
}

func (silver SilverPolishServerConnection) Polish(input []byte) []byte {
	var output []byte

	// Generate random nonce
	nonce := make([]byte, silver.polishCipher.NonceSize())
	rand.Read(nonce)

	silver.polishCipher.Seal(output, nonce, input, nil)

	result := append(nonce, output...)

	return result
}

func (silver SilverPolishServerConnection) Unpolish(input []byte) []byte {
	var output []byte

	nonceSize := silver.polishCipher.NonceSize()
	nonce := input[:nonceSize]
	data := input[nonceSize:]

	silver.polishCipher.Open(output, nonce, data, nil)

	return output
}
