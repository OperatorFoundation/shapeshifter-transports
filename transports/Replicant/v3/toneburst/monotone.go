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

package toneburst

import (
	"errors"
	"github.com/OperatorFoundation/monolith-go/monolith"
	"net"
)

type MonotoneConfig struct {
	AddSequences    *monolith.Instance
	RemoveSequences *monolith.Description
	SpeakFirst      bool
}

func (config MonotoneConfig) Construct() (ToneBurst, error) {
	return NewMonotone(config), nil
}

type Monotone struct {
	config MonotoneConfig
	buffer *monolith.Buffer
	context *monolith.Context
}

func NewMonotone(config MonotoneConfig) *Monotone {
	buffer := monolith.NewEmptyBuffer()
	context := monolith.NewEmptyContext()

	return &Monotone{
		config:  config,
		buffer:  buffer,
		context: context,
	}
}

//TODO: Implement Perform
func (monotone *Monotone) Perform(conn net.Conn) error {

	var addMessages []monolith.Message
	var removeParts []monolith.Monolith

	if monotone.config.AddSequences != nil {
		addMessages = monotone.config.AddSequences.Messages()
	}

	if monotone.config.RemoveSequences != nil {
		removeParts = monotone.config.RemoveSequences.Parts
	}

	if monotone.config.SpeakFirst {
		if addMessages == nil || len(addMessages) < 1 {
			println("Invalid configuration, cannot speak first when there is nothing to add.")
			return errors.New("invalid configuration, cannot speak first when there is nothing to add")
		}

		//Get the first sequence in the list of add sequences
		firstMessage := addMessages[0]
		addMessages = addMessages[1:]
		addBytes := firstMessage.Bytes()

		writeError := writeAll(conn, addBytes)
		if writeError != nil {
			return writeError
		}
	}

	for {
		if (removeParts == nil || len(removeParts) < 1) && (addMessages == nil || len(addMessages) < 1) {
			return nil
		}

		if removeParts != nil && len(removeParts) > 0 {
			removePart := removeParts[0]
			removeParts = removeParts[1:]

			validated, readAllError := monotone.readAll(conn, removePart)
			if readAllError != nil {
				println("Error reading data: ", readAllError.Error())
				return readAllError
			}

			if !validated {
				return errors.New("failed to validate toneburst data, invalid remove sequence")
			}
		}

		if addMessages != nil && len(addMessages) > 0 {
			//Get the first sequence in the list of add sequences
			firstMessage := addMessages[0]
			addMessages = addMessages[1:]
			addBytes := firstMessage.Bytes()

			writeError := writeAll(conn, addBytes)
			if writeError != nil {
				return writeError
			}
		}
	}
}

func (monotone Monotone) readAll(conn net.Conn, part monolith.Monolith) (bool, error) {
	receivedData := make([]byte, part.Count())
	_, readError := conn.Read(receivedData)
	if readError != nil {
		println("Received an error while trying to receive data: ", readError.Error())
		return false, readError
	}

	monotone.buffer.Push(receivedData)
	validated := part.Validate(monotone.buffer, monotone.context)

	switch validated {

	case monolith.Valid:
		return true, nil
	case monolith.Invalid:
		println("Failed to validate the received data.")
		return false, errors.New("failed to validate the received data")
	case monolith.Incomplete:
		println("Failed to validate the received data, data was incomplete.")
		return false, errors.New("failed to validate the received data, data was incomplete")
	default:
		println("Validate returned an unknown value.")
		return false, errors.New("validate returned an unknown value")
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
