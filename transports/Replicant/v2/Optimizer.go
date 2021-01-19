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

package replicant

import (
	"golang.org/x/net/proxy"
	"net"
)

// This makes Replicant compliant with Optimizer
type Transport struct {
	Config  ClientConfig
	Address string
	Dialer  proxy.Dialer
}

// TODO: the dial we call currently does not return an error
func (transport Transport) Dial() (net.Conn, error) {
	conn, dialErr := transport.Dialer.Dial("tcp", transport.Address)
	if dialErr != nil {
		return nil, dialErr
	}

	dialConn := conn
	transportConn, err := NewClientConnection(conn, transport.Config)
	if err != nil {
		_ = dialConn.Close()
		return nil, err
	}

	return transportConn, nil

	//replicantTransport := New(transport.Config, transport.Dialer)
	//conn := replicantTransport.Dial(transport.Address)
	//conn, err:= replicantTransport.Dial(transport.Address), errors.New("connection failed")
	//if err != nil {
	//	return nil, err
	//} else {
	//	return conn, nil
	//}
	//return conn, nil
}

