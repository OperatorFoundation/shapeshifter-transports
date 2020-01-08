package toneburst

import (
	"errors"
	"github.com/OperatorFoundation/monolith-go/monolith"
	"net"
)

type MonotoneConfig struct {
	AddSequences    []monolith.Instance
	RemoveSequences []monolith.Description
	SpeakFirst      bool
}

type Monotone struct {
	config MonotoneConfig
}

func NewMonotone(config MonotoneConfig) *Monotone {
	return &Monotone{config: config}
}

//TODO: Implement Perform
func (monotone *Monotone) Perform(conn net.Conn) error {

	if monotone.config.SpeakFirst {
		if len(monotone.config.AddSequences) == 0 {
			println("Invalid configuration, cannot speak first when there is nothing to add.")
			return errors.New("invalid configuration, cannot speak first when there is nothing to add")
		}

		//Get the first sequence in the list of add sequences
		addInstance := monotone.config.AddSequences[0]
		addBytes := addInstance.Bytes()

		// Remove this sequence from the list
		monotone.config.AddSequences = monotone.config.AddSequences[1:]

		writeError := writeAll(conn, addBytes)
		if writeError != nil {
			return writeError
		}
	}

	receiveDataBuffer := make([]byte, 0)
	for {
		if len(monotone.config.RemoveSequences) == 0 {
			return nil
		}

		removeSequenceDescription := monotone.config.RemoveSequences[0]
		monotone.config.RemoveSequences = monotone.config.RemoveSequences[:1]

		dataBuffer, readAllError := readAll(conn, removeSequenceDescription, receiveDataBuffer)
		if readAllError != nil {
			println("Error reading data: ", readAllError)
			return readAllError
		}
		receiveDataBuffer = dataBuffer

		if len(monotone.config.AddSequences) == 0 {
			return nil
		}

		//Get the first sequence in the list of add sequences
		addInstance := monotone.config.AddSequences[0]

		// Remove this sequence from the list
		monotone.config.AddSequences = monotone.config.AddSequences[1:]

		writeError := writeAll(conn, addInstance.Bytes())
		if writeError != nil {
			return writeError
		}
	}
}

func readAll(conn net.Conn, removeSequenceDescription monolith.Description, receiveDataBuffer []byte) ([]byte, error) {
	receivedData := make([]byte, 0)
	_, readError := conn.Read(receivedData)
	if readError != nil {
		println("Received an error while trying to receive data: ", readError.Error())
		return nil, readError
	}

	receiveDataBuffer = append(receiveDataBuffer, receivedData...)

	remainingData, validated := removeSequenceDescription.Validate(receivedData)
	for validated == monolith.Incomplete {
		_, readError = conn.Read(receivedData)
		if readError != nil {
			println("Received an error while trying to receive data: ", readError.Error())
			return nil, readError
		}

		println("Attempting to read data from connection. Received data length: ", len(receivedData))
		receiveDataBuffer = append(receiveDataBuffer, receivedData...)
		remainingData, validated = removeSequenceDescription.Validate(receivedData)
	}

	switch validated {

	case monolith.Valid:
		return remainingData, nil
	case monolith.Invalid:
		println("Failed to validate the received data.")
		return nil, errors.New("failed to validate the received data")
	default:
		println("Validate returned an unknown value.")
		return nil, errors.New("validate returned an unknown value")
	}
}

func writeAll(conn net.Conn, addBytes []byte) error {
	writtenCount, writeError := conn.Write(addBytes)
	if writeError != nil {
		println("Received an error while attempting to write data: ", writeError)
		return writeError
	}

	for writtenCount < len(addBytes) {
		addBytes = addBytes[writtenCount:]
		writtenCount, writeError = conn.Write(addBytes)
		if writeError != nil {
			println("Received an error while attempting to write data: ", writeError)
			return writeError
		}
	}

	return nil
}
