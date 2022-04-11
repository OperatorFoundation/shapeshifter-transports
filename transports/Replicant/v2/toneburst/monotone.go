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
	"bufio"
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
	Config  MonotoneConfig
	Buffer  *monolith.Buffer
	Context *monolith.Context
}

func NewMonotone(config MonotoneConfig) *Monotone {
	buffer := monolith.NewEmptyBuffer()
	context := monolith.NewEmptyContext()

	return &Monotone{
		Config:  config,
		Buffer:  buffer,
		Context: context,
	}
}

//TODO: Implement Perform
func (monotone *Monotone) Perform(conn net.Conn) error {

	var addMessages []monolith.Message
	var removeParts []monolith.Monolith

	if monotone.Config.AddSequences != nil {
		addMessages = monotone.Config.AddSequences.Messages()
	}

	if monotone.Config.RemoveSequences != nil {
		removeParts = monotone.Config.RemoveSequences.Parts
	}

	if monotone.Config.SpeakFirst {
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
	switch partType := part.(type) {
	case monolith.BytesPart:
		receivedData := make([]byte, partType.Count())
		_, readError := conn.Read(receivedData)
		if readError != nil {
			println("Received an error while trying to receive data: ", readError.Error())
			return false, readError
		}

		monotone.Buffer.Push(receivedData)
		validated := part.Validate(monotone.Buffer, monotone.Context)

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

	case monolith.StringsPart:
		for _, item := range partType.Items {
			var receivedBuffer []byte
			switch stringsType := item.(type) {
			case monolith.VariableStringType:
				receivedString, readError := bufio.NewReader(conn).ReadString(stringsType.EndDelimiter)
				if readError != nil {
					return false, readError
				}

				receivedBuffer = []byte(receivedString)
			case monolith.FixedStringType:
				receivedBuffer = make([]byte, stringsType.Count())
				_, readError := conn.Read(receivedBuffer)
				if readError != nil {
					println("Received an error while trying to receive data: ", readError.Error())
					return false, readError
				}
			}

			monotone.Buffer.Push(receivedBuffer)
			validated := item.Validate(monotone.Buffer, monotone.Context)

			switch validated {
			case monolith.Valid:
				continue
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
	}

	return true, nil
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
