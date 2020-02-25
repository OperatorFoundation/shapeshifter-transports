package meekserver

import (
	"net"
	"net/http"
	"testing"
)

func checkExpected(t *testing.T, req *http.Request, expected net.IP) {
	ip, err := originalClientIP(req)
	// If ip is nil, then err must not be nil; and if ip is not nil, then
	// err must be nil.
	if (ip == nil && err == nil) || (ip != nil && err != nil) {
		t.Errorf("%v returned ip=%v and err=%v", req, ip, expected)
		return
	}
	// Either both are nil or ip is equal to expected.
	if (ip == nil && expected != nil) ||
		(ip != nil && !ip.Equal(expected)) {
		t.Errorf("%v got %v, expected %v", req, ip, expected)
	}
}

// Test that in the absence of headers indicating the original client IP
// address, originalClientIP returns an answer based on RemoteAddr.
func TestOriginalClientIPRemoteAddr(t *testing.T) {
	tests := []struct {
		RemoteAddr string
		Expected   net.IP
	}{
		{"", nil},
		{"1.2.3.4:1234", net.IPv4(1, 2, 3, 4)},
		{"[1:2::3:4]:1234", net.IP{0, 1, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 4}},
		// Bad syntax.
		{"1.2.3.4", nil},
		{"1:2::3:4:1234", nil},
		{"xyz", nil},
	}

	for _, test := range tests {
		req := &http.Request{
			RemoteAddr: test.RemoteAddr,
		}
		checkExpected(t, req, test.Expected)
	}
}

// Test that originalClientIP reads the Meek-IP and X-Forwarded-For headers if
// present.
func TestOriginalClientXForwardedFor(t *testing.T) {
	tests := []struct {
		XForwardedFor string
		Expected      net.IP
	}{
		{"", nil},
		{"1.2.3.4", net.IPv4(1, 2, 3, 4)},
		{"1:2::3:4", net.IP{0, 1, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 4}},
		{"1.2.3.4,1:2::3:4", net.IPv4(1, 2, 3, 4)},
		{"1.2.3.4, 1:2::3:4", net.IPv4(1, 2, 3, 4)},
		{"1:2::3:4, 1.2.3.4", net.IP{0, 1, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 4}},
		// Bad syntax.
		{"xyz", nil},
		{"1.2.3.4:1234", nil},
		// Only try to parse the first element
		{"1.2.3.4, xyz", net.IPv4(1, 2, 3, 4)},
		{"1.2.3.4,,,", net.IPv4(1, 2, 3, 4)},
		{"xyz, 1.2.3.4", nil},
		// X-Forwarded-For doesn't use square brackets on IPv6.
		{"[1:2::3:4]", nil},
		{"[1:2::3:4]:1234", nil},
	}

	for _, test := range tests {
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Set("Meek-IP", test.XForwardedFor)
		checkExpected(t, req, test.Expected)
		req.Header.Set("X-Forwarded-For", test.XForwardedFor)
		checkExpected(t, req, test.Expected)
	}
}

// Test that headers prevent reading RemoteAddr, even if the headers cannot be
// parsed.
func TestOriginalClientPrecedence(t *testing.T) {
	tests := []struct {
		Req      http.Request
		Expected net.IP
	}{
		{
			http.Request{},
			nil,
		},
		{
			http.Request{
				RemoteAddr: "5.6.7.8:5678",
			},
			net.IPv4(5, 6, 7, 8),
		},
		{
			http.Request{
				RemoteAddr: "5.6.7.8:5678",
				Header: http.Header{
					http.CanonicalHeaderKey("Meek-IP"): []string{"1.2.3.4"},
				},
			},
			net.IPv4(1, 2, 3, 4),
		},
		{
			http.Request{
				RemoteAddr: "5.6.7.8:5678",
				Header: http.Header{
					http.CanonicalHeaderKey("Meek-IP"): []string{"1:2::3:4"},
				},
			},
			net.IP{0, 1, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 4},
		},
		{
			http.Request{
				RemoteAddr: "5.6.7.8:5678",
				Header: http.Header{
					http.CanonicalHeaderKey("X-Forwarded-For"): []string{"1.2.3.4"},
				},
			},
			net.IPv4(1, 2, 3, 4),
		},
		{
			http.Request{
				RemoteAddr: "5.6.7.8:5678",
				Header: http.Header{
					http.CanonicalHeaderKey("X-Forwarded-For"): []string{"1:2::3:4"},
				},
			},
			net.IP{0, 1, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 4},
		},
		// Meek-IP has precedence over X-Forwarded-For if both are set.
		{
			http.Request{
				RemoteAddr: "5.6.7.8:5678",
				Header: http.Header{
					http.CanonicalHeaderKey("Meek-IP"):         []string{"1.2.3.4"},
					http.CanonicalHeaderKey("X-Forwarded-For"): []string{"2.2.2.2"},
				},
			},
			net.IPv4(1, 2, 3, 4),
		},
		// X-Forwarded-For shadows RemoteAddr, even if bad syntax.
		{
			http.Request{
				RemoteAddr: "5.6.7.8:5678",
				Header: http.Header{
					http.CanonicalHeaderKey("X-Forwarded-For"): []string{"xyz"},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		checkExpected(t, &test.Req, test.Expected)
	}
}
