package meekserver

import (
	"flag"
	pt "git.torproject.org/pluggable-transports/goptlib.git"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

//undefined: startListener and startListenerTLS
func TestMeekServerListen(t *testing.T) {
	meekserverTransport := meekServer{true, "cert", "key" }
	listener := meekserverTransport.Listen("127.0.0.1:1234")
	if listener == nil {
		t.Fail()
	}
}

//The --cert and --key options are required
func TestMeekServerMain(t *testing.T) {
	var disableTLS bool
	var certFilename, keyFilename string
	var logFilename string
	var port int

	flag.BoolVar(&disableTLS, "disable-tls", false, "don't use HTTPS")
	flag.StringVar(&certFilename, "cert", "", "TLS certificate file (required without --disable-tls)")
	flag.StringVar(&keyFilename, "key", "", "TLS private key file (required without --disable-tls)")
	flag.StringVar(&logFilename, "log", "", "name of log file")
	flag.IntVar(&port, "port", 0, "port to listen on")
	flag.Parse()

	if logFilename != "" {
		f, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("error opening log file: %s", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	if disableTLS {
		if certFilename != "" || keyFilename != "" {
			log.Fatalf("The --cert and --key options are not allowed with --disable-tls.\n")
		}
	} else {
		if certFilename == "" || keyFilename == "" {
			log.Fatalf("The --cert and --key options are required.\n")
		}
	}

	var err error
	ptInfo, err = pt.ServerSetup([]string{ptMethodName})
	if err != nil {
		log.Fatalf("error in ServerSetup: %s", err)
	}

	log.Printf("starting")
	listeners := make([]net.Listener, 0)
	for _, bindaddr := range ptInfo.Bindaddrs {
		if port != 0 {
			bindaddr.Addr.Port = port
		}
		switch bindaddr.MethodName {
		case ptMethodName:
			var ln net.Listener
			if disableTLS {
				ln, err = startListener("tcp", bindaddr.Addr)
			} else {
				ln, err = startListenerTLS("tcp", bindaddr.Addr, certFilename, keyFilename)
			}
			if err != nil {
				pt.SmethodError(bindaddr.MethodName, err.Error())
				break
			}
			pt.Smethod(bindaddr.MethodName, ln.Addr())
			listeners = append(listeners, ln)
		default:
			pt.SmethodError(bindaddr.MethodName, "no such method")
		}
	}
	pt.SmethodsDone()

	var numHandlers int = 0
	var sig os.Signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for first signal.
	sig = nil
	for sig == nil {
		select {
		case n := <-handlerChan:
			numHandlers += n
		case sig = <-sigChan:
			log.Printf("got signal %s", sig)
		}
	}
	for _, ln := range listeners {
		ln.Close()
	}

	if sig == syscall.SIGTERM {
		log.Printf("done")
		return
	}

	// Wait for second signal or no more handlers.
	sig = nil
	for sig == nil && numHandlers != 0 {
		select {
		case n := <-handlerChan:
			numHandlers += n
		case sig = <-sigChan:
			log.Printf("got second signal %s", sig)
		}
	}

	log.Printf("done")
}

