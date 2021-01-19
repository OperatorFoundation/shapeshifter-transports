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
	"os"
	"runtime"
	"testing"
)

const data = "test"

func TestMain(m *testing.M) {
	_ = RunLocalObfs4Server("test")
	_ = RunLocalObfs4ServerFactory("test")
	os.Exit(m.Run())
}

func TestObfs4(t *testing.T) {
	print(runtime.GOOS)
	clientConfig, launchErr := RunObfs4Client()
	if launchErr != nil {
		t.Fail()
		return
	}
	//create client buffer
	clientBuffer := make([]byte, 4)
	//call dial on client and check error
	clientConn, dialErr := clientConfig.Dial("127.0.0.1:1234")
	if dialErr != nil {
		t.Fail()
		return
	}

	//write data from clientConn for server to read
	_, clientWriteErr := clientConn.Write([]byte(data))
	if clientWriteErr != nil {
		t.Fail()
		return
	}

	//read on client side
	_, clientReadErr := clientConn.Read(clientBuffer)
	if clientReadErr != nil {
		t.Fail()
		return
	}
}

func TestObfs4Factory(t *testing.T) {
	print(runtime.GOOS)
	clientConfig, launchErr := RunObfs4ClientFactory()
	if launchErr != nil {
		t.Fail()
		return
	}
	//create client buffer
	clientBuffer := make([]byte, 4)
	//call dial on client and check error
	clientConn, dialErr := clientConfig.Dial()
	if dialErr != nil {
		t.Fail()
		return
	}

	//write data from clientConn for server to read
	_, clientWriteErr := clientConn.Write([]byte(data))
	if clientWriteErr != nil {
		t.Fail()
		return
	}

	//read on client side
	_, clientReadErr := clientConn.Read(clientBuffer)
	if clientReadErr != nil {
		t.Fail()
		return
	}
}