// certificate.go - Certificate management for meek-server.

// +build go1.6

package meekserver

import (
	"crypto/tls"
	"log"
	"os"
	"sync"
	"time"
)

const certLoadErrorRateLimit = 1 * time.Minute

type certContext struct {
	sync.Mutex

	certFile string
	keyFile  string

	certFileInfo os.FileInfo
	keyFileInfo  os.FileInfo
	cachedCert   *tls.Certificate

	lastWarnAt time.Time
}

func newCertContext(certFilename, keyFilename string) (*certContext, error) {
	ctx := new(certContext)
	ctx.certFile = certFilename
	ctx.keyFile = keyFilename
	if _, err := ctx.reloadCertificate(); err != nil {
		return nil, err
	}
	return ctx, nil
}

func (ctx *certContext) reloadCertificate() (*tls.Certificate, error) {
	doReload := true

	// XXX/Yawning: I assume compared to everything else related to TLS
	// handshakes, stat() is cheap.  If not, ratelimit here.  gettimeofday()
	// is vDSO-ed so it would be significantly faster than the syscalls.

	var err error
	var cfInfo, kfInfo os.FileInfo
	if cfInfo, err = os.Stat(ctx.certFile); err == nil {
		kfInfo, err = os.Stat(ctx.keyFile)
	}

	ctx.Lock()
	defer ctx.Unlock()

	// Grab the cached certificate, compare the modification times if able.
	cert := ctx.cachedCert
	if err != nil {
		// If stat fails, we likely aren't going to be able to reload, so
		// return early.
		return cert, err
	} else if ctx.cachedCert != nil {
		// Only compare the file times if there's actually a cached cert,
		// and reload the cert if either the key or the certificate have
		// been modified.
		doReload = !ctx.certFileInfo.ModTime().Equal(cfInfo.ModTime()) || !ctx.keyFileInfo.ModTime().Equal(kfInfo.ModTime())
	}

	// Attempt to load the updated certificate, if required.
	if doReload {
		newCert, err := tls.LoadX509KeyPair(ctx.certFile, ctx.keyFile)
		if err != nil {
			// If the load fails, return the old certificate, so that it can
			// be used till the load succeeds.
			return cert, err
		}

		// If the user regenerates the cert/key between the stat() and
		// LoadX509KeyPair calls, this will race, but will self-correct
		// after the next reloadCertificate() call because doReload will
		// be true.

		ctx.cachedCert = &newCert
		ctx.certFileInfo = cfInfo
		ctx.keyFileInfo = kfInfo

		cert = ctx.cachedCert
	}
	return cert, nil
}

func (ctx *certContext) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := ctx.reloadCertificate()
	if err != nil {
		// Failure to reload the certificate is a non-fatal error as this
		// may be a filesystem related race condition.  There is nothing
		// preventing the next callback from hopefully succeeding, so rate
		// limit an error log.
		now := time.Now()
		if now.After(ctx.lastWarnAt.Add(certLoadErrorRateLimit)) {
			ctx.lastWarnAt = now
			log.Printf("failed to reload certificate: %v", err)
		}
	}

	// This should NEVER happen because we will continue to use the old
	// certificate on load failure, and we will never be calling the
	// listener GetCertificate() callback if the initial load fails.
	if cert == nil {
		panic("no cached certificate available")
	}

	return cert, nil
}
