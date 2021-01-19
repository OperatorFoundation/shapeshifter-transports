package meekserver

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// openssl genpkey -out key1.pem -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -pkeyopt ec_param_enc:named_curve
const key1PEM = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgwBuadiHWuhiCwUHK
pvzM0MYVySH8FE5T/76qoDBozzChRANCAAQao88iJDj284z5YsSOwynblbNfhRHL
dhInn3bl8dYUr0s88q7yaOHW0riAYAX8Q/G5zoP/P1MgNPFMYV76eSY4
-----END PRIVATE KEY-----`

// openssl req -out cert1.pem -x509 -new -nodes -key key1.pem -days 1 -subj "/CN=meek-server.example.com"
const cert1PEM = `-----BEGIN CERTIFICATE-----
MIIBjjCCATSgAwIBAgIJAKpvM1Hu/AeyMAoGCCqGSM49BAMCMCIxIDAeBgNVBAMM
F21lZWstc2VydmVyLmV4YW1wbGUuY29tMB4XDTE3MDMyMTIyNTM0N1oXDTE3MDMy
MjIyNTM0N1owIjEgMB4GA1UEAwwXbWVlay1zZXJ2ZXIuZXhhbXBsZS5jb20wWTAT
BgcqhkjOPQIBBggqhkjOPQMBBwNCAAQao88iJDj284z5YsSOwynblbNfhRHLdhIn
n3bl8dYUr0s88q7yaOHW0riAYAX8Q/G5zoP/P1MgNPFMYV76eSY4o1MwUTAdBgNV
HQ4EFgQU/FIFX5DX58BFhNBSWV0ulWmS+XIwHwYDVR0jBBgwFoAU/FIFX5DX58BF
hNBSWV0ulWmS+XIwDwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNIADBFAiEA
mylhVbnAd0KEQoaIEH1whj9oUxlk2kWU5G8daG5uUjMCIBW6fwv0cbYmyzspCMqJ
eib1vgFnhGTI44K05cunpXJ+
-----END CERTIFICATE-----`

// openssl genpkey -out key2.pem -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -pkeyopt ec_param_enc:named_curve
const key2PEM = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgUbNm7exkEPKMv/DT
B3UMv6XevnD2E99W3Rp8L/a1v12hRANCAATqnza9H9uNa8EvoKm2GOLvw25kR7OA
oVHZiyaXdBeB480FBUtRmWUukLZFxp/QStd4OCwaOwWGtXGlspM2LEum
-----END PRIVATE KEY-----`

// openssl req -out cert2.pem -x509 -new -nodes -key key2.pem -days 1 -subj "/CN=meek-server.example.com"
const cert2PEM = `-----BEGIN CERTIFICATE-----
MIIBjzCCATSgAwIBAgIJAKvKZ4vqwpySMAoGCCqGSM49BAMCMCIxIDAeBgNVBAMM
F21lZWstc2VydmVyLmV4YW1wbGUuY29tMB4XDTE3MDMyMTIyNTQzN1oXDTE3MDMy
MjIyNTQzN1owIjEgMB4GA1UEAwwXbWVlay1zZXJ2ZXIuZXhhbXBsZS5jb20wWTAT
BgcqhkjOPQIBBggqhkjOPQMBBwNCAATqnza9H9uNa8EvoKm2GOLvw25kR7OAoVHZ
iyaXdBeB480FBUtRmWUukLZFxp/QStd4OCwaOwWGtXGlspM2LEumo1MwUTAdBgNV
HQ4EFgQUdqTMumWa7f965k/SLgWJT0tIlcswHwYDVR0jBBgwFoAUdqTMumWa7f96
5k/SLgWJT0tIlcswDwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNJADBGAiEA
yOhMqBPZLpqnjTGD1OgOc1N1SkS53bdWAazAZgLgwvICIQC7hFOK1j74Frh+1l3h
kLn8Jbjpp6jzNBhQhVHF/nj3CQ==
-----END CERTIFICATE-----`

const badSyntax = "hello world\n"

// Backdate file times so any modifications will make them appear different;
// otherwise immediately overwriting the files may happen quickly enough that
// the times don't change.
func backdateFile(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	timestamp := info.ModTime().Add(-10 * time.Second)
	return os.Chtimes(filename, timestamp, timestamp)
}

func mustBackdateFile(filename string) {
	err := backdateFile(filename)
	if err != nil {
		panic(err)
	}
}

// Return the path to a newly created temporary file with the given contents.
func makeTempFileFromContents(contents []byte) (string, error) {
	f, err := ioutil.TempFile("", "meek-server-certificate-test-")
	if err != nil {
		return "", err
	}

	_, err = f.Write(contents)
	if err != nil {
		f.Close()
		return f.Name(), err
	}

	err = backdateFile(f.Name())
	f.Close()
	return f.Name(), err
}

func mustMakeTempFileFromContents(contents []byte) string {
	f, err := makeTempFileFromContents(contents)
	if err != nil {
		panic(err)
	}
	return f
}

// Return a random filename that is unlikely to exist.
func makeNonexistentFilename() string {
	fileBytes := make([]byte, 8)
	_, err := rand.Read(fileBytes)
	if err != nil {
		panic(err)
	}
	return filepath.Join(os.TempDir(), hex.EncodeToString(fileBytes))
}

// Call tls.X509KeyPair and panic if it fails.
func mustLoadCertificate(certPEM, keyPEM []byte) *tls.Certificate {
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &cert
}

func mustWriteFile(filename string, data []byte) {
	err := ioutil.WriteFile(filename, data, 0600)
	if err != nil {
		panic(err)
	}
}

// A bunch of temporary certificate filenames and the like.
type testFiles struct {
	cert1 *tls.Certificate
	cert2 *tls.Certificate

	key1Filename        string
	key2Filename        string
	cert1Filename       string
	cert2Filename       string
	badSyntaxFilename   string
	nonexistentFilename string
}

func loadTestFiles() *testFiles {
	var files testFiles

	files.cert1 = mustLoadCertificate([]byte(cert1PEM), []byte(key1PEM))
	files.cert2 = mustLoadCertificate([]byte(cert2PEM), []byte(key2PEM))

	files.key1Filename = mustMakeTempFileFromContents([]byte(key1PEM))
	files.key2Filename = mustMakeTempFileFromContents([]byte(key2PEM))
	files.cert1Filename = mustMakeTempFileFromContents([]byte(cert1PEM))
	files.cert2Filename = mustMakeTempFileFromContents([]byte(cert2PEM))
	files.badSyntaxFilename = mustMakeTempFileFromContents([]byte(badSyntax))
	files.nonexistentFilename = makeNonexistentFilename()

	return &files
}

// Delete temporary files created by loadTestFiles (to be called in a defer
// handler).
func (files *testFiles) Cleanup() {
	_ = os.Remove(files.key1Filename)
	_ = os.Remove(files.key2Filename)
	_ = os.Remove(files.cert1Filename)
	_ = os.Remove(files.cert2Filename)
	_ = os.Remove(files.badSyntaxFilename)
}

// Check if two certificate chains are equal.
func certificatesEqual(cert1, cert2 *tls.Certificate) bool {
	if len(cert1.Certificate) != len(cert2.Certificate) {
		return false
	}
	for i := range cert1.Certificate {
		if !bytes.Equal(cert1.Certificate[i], cert2.Certificate[i]) {
			return false
		}
	}
	return true
}

// Call ctx.reloadCertificate and check if the certificate and error status are
// as expected.
func checkCertificate(t *testing.T, ctx *certContext, expectedCert *tls.Certificate, expectedError bool) {
	cert, err := ctx.reloadCertificate()
	if expectedError && err == nil {
		t.Errorf("expected error, got %v\n", err)
	} else if !expectedError && err != nil {
		t.Errorf("expected no error, got %v\n", err)
	}
	if !certificatesEqual(cert, expectedCert) {
		t.Errorf("certificate was other than expected")
	}
}

func TestNewCertContext(t *testing.T) {
	files := loadTestFiles()
	defer files.Cleanup()

	var ctx *certContext
	var err error

	// Test with one or both files nonexistent.
	ctx, err = newCertContext(files.nonexistentFilename, files.nonexistentFilename)
	if err == nil {
		t.Errorf("did not raise error on nonexistent cert and key")
	}
	ctx, err = newCertContext(files.cert1Filename, files.nonexistentFilename)
	if err == nil {
		t.Errorf("did not raise error on nonexistent cert")
	}
	ctx, err = newCertContext(files.nonexistentFilename, files.key1Filename)
	if err == nil {
		t.Errorf("did not raise error on nonexistent key")
	}

	// Test with bad syntax.
	ctx, err = newCertContext(files.badSyntaxFilename, files.badSyntaxFilename)
	if err == nil {
		t.Errorf("did not raise error on bad-syntax cert and key")
	}
	ctx, err = newCertContext(files.cert1Filename, files.badSyntaxFilename)
	if err == nil {
		t.Errorf("did not raise error on bad-syntax cert")
	}
	ctx, err = newCertContext(files.badSyntaxFilename, files.key1Filename)
	if err == nil {
		t.Errorf("did not raise error on bad-syntax key")
	}

	// Test with certificate and key that don't match.
	ctx, err = newCertContext(files.cert1Filename, files.key2Filename)
	if err == nil {
		t.Errorf("did not raise error on mismatched cert and key")
	}
	ctx, err = newCertContext(files.cert2Filename, files.key1Filename)
	if err == nil {
		t.Errorf("did not raise error on mismatched cert and key")
	}

	// Test with everything good.
	ctx, err = newCertContext(files.cert1Filename, files.key1Filename)
	if ctx == nil || err != nil {
		t.Fatalf("raised an error: %s", err)
	}
}

// Test that reloadCertificate continues returning the old certificate if files
// are deleted.
func TestDelete(t *testing.T) {
	files := loadTestFiles()
	defer files.Cleanup()

	var ctx *certContext
	var err error

	ctx, err = newCertContext(files.cert1Filename, files.key1Filename)
	if ctx == nil || err != nil {
		t.Fatalf("raised an error: %s", err)
	}
	checkCertificate(t, ctx, files.cert1, false)
	// Try removing the cert file; cert should be the same but now raise an error.
	_ = os.Remove(files.cert1Filename)
	checkCertificate(t, ctx, files.cert1, true)

	ctx, err = newCertContext(files.cert2Filename, files.key2Filename)
	if ctx == nil || err != nil {
		t.Fatalf("raised an error: %s", err)
	}
	checkCertificate(t, ctx, files.cert2, false)
	// Try removing the key file; cert should be the same but now raise an error.
	_ = os.Remove(files.key2Filename)
	checkCertificate(t, ctx, files.cert2, true)
}

// Test replacing the contents of cert and key files.
func TestReplace(t *testing.T) {
	files := loadTestFiles()
	defer files.Cleanup()

	var ctx *certContext
	var err error

	ctx, err = newCertContext(files.cert1Filename, files.key1Filename)
	if ctx == nil || err != nil {
		t.Fatalf("raised an error: %s", err)
	}
	checkCertificate(t, ctx, files.cert1, false)

	// Replace cert file with junk.
	mustWriteFile(files.cert1Filename, []byte(badSyntax))
	checkCertificate(t, ctx, files.cert1, true)
	// Put it back to normal.
	mustWriteFile(files.cert1Filename, []byte(cert1PEM))
	checkCertificate(t, ctx, files.cert1, false)

	// Replace key file with junk.
	mustWriteFile(files.key1Filename, []byte(badSyntax))
	checkCertificate(t, ctx, files.cert1, true)
	// Put it back to normal.
	mustWriteFile(files.key1Filename, []byte(key1PEM))
	checkCertificate(t, ctx, files.cert1, false)

	mustBackdateFile(files.cert1Filename)
	mustBackdateFile(files.key1Filename)
	checkCertificate(t, ctx, files.cert1, false)

	// Replace cert1 with cert2 contents; expect to still get cert1, with an error.
	mustWriteFile(files.cert1Filename, []byte(cert2PEM))
	checkCertificate(t, ctx, files.cert1, true)
	// Replace key1 with key2 contents; now we expect to be using cert2.
	mustWriteFile(files.key1Filename, []byte(key2PEM))
	checkCertificate(t, ctx, files.cert2, false)
}
