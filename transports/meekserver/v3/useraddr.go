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

package meekserver

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// Get the first hostname in an X-Forwarded-For header field value.
//
// The X-Forwarded-For header is set automatically by CDNs (and other proxies).
// There doesn't seem to be any formal specification of it.
// https://en.wikipedia.org/wiki/X-Forwarded-For says it was originally
// implemented in Squid. The format appears to be: one or more IPv4 or IPv6
// addresses, without port numbers, separated by commas and optional whitespace.
// Because the format is unspecified, we only look at whatever is before the
// first comma (or end of string).
func getXForwardedFor(xForwardedFor string) (string, error) {
	parts := strings.SplitN(xForwardedFor, ",", 2)
	if len(parts) < 1 {
		return "", fmt.Errorf("bad X-Forwarded-For value %q", xForwardedFor)
	}
	return strings.TrimSpace(parts[0]), nil
}

// Return the original client IP address as best as it can be determined.
func originalClientIP(req *http.Request) (net.IP, error) {
	var host string
	var err error

	xForwardedFor := req.Header.Get("Meek-IP")
	if xForwardedFor == "" {
		xForwardedFor = req.Header.Get("X-Forwarded-For")
	}
	if xForwardedFor != "" {
		host, err = getXForwardedFor(xForwardedFor)
	} else {
		host, _, err = net.SplitHostPort(req.RemoteAddr)
	}
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("cannot parse %q as IP address", host)
	}
	return ip, nil
}
