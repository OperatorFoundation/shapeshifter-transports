/*
 * Copyright (c) 2015, Yawning Angel <yawning at torproject dot org>
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

// Package meeklite provides an implementation of the Meek circumvention
// protocol.  Only a client implementation is provided, and no effort is
// made to normalize the TLS fingerprint.
//
// It borrows quite liberally from the real meek-client code.

package meeklite

import (
	"github.com/kataras/golog"
	"golang.org/x/net/proxy"
	gourl "net/url"
	"os"
	"testing"
)

const data = "test"

func TestMeeklite(t *testing.T) {
	MakeLog()
	//create a server
	config := NewMeekTransportWithFront("https://transport-canary-meek.appspot.com/", "www.google.com", proxy.Direct)

	//create client buffer
	clientBuffer := make([]byte, 4)
	//call dial on client and check error
	clientConn, connErr := config.Dial()
	if connErr != nil {
		t.Fail()
		return
	}

	//write data from clientConn for server to read
	writeBytes, clientWriteErr := clientConn.Write([]byte(data))
	if clientWriteErr != nil {
		t.Fail()
		return
	}
	if writeBytes <= 0 {
		t.Fail()
		return
	}

	//read on client side
	readBytes, clientReadErr := clientConn.Read(clientBuffer)
	if clientReadErr != nil {
		t.Fail()
		return
	}
	if readBytes <= 0 {
		t.Fail()
		return
	}
}

func TestFactoryMeeklite(t *testing.T) {
	MakeLog()
	//create a server
	urlString := "https://transport-canary-meek.appspot.com/"
	Url, err := gourl.Parse(urlString)
	if err != nil {
		t.Fail()
		return
	}

	config := NewMeekFactoryTransportWithFront(Url, "www.google.com", proxy.Direct)

	//create client buffer
	clientBuffer := make([]byte, 4)
	//call dial on client and check error
	clientConn, connErr := config.Dial()
	if connErr != nil {
		t.Fail()
		return
	}

	//write data from clientConn for server to read
	writeBytes, clientWriteErr := clientConn.Write([]byte(data))
	if clientWriteErr != nil {
		t.Fail()
		return
	}
	if writeBytes <= 0 {
		t.Fail()
		return
	}

	//read on client side
	readBytes, clientReadErr := clientConn.Read(clientBuffer)
	if clientReadErr != nil {
		t.Fail()
		return
	}
	if readBytes <= 0 {
		t.Fail()
		return
	}
}

func MakeLog() {
	golog.SetLevel("debug")
	golog.SetOutput(os.Stderr)
}