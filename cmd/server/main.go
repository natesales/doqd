package main

import (
	"crypto/tls"
	"flag"
	"os"
	"runtime"

	"github.com/natesales/doq/pkg/server"
	log "github.com/sirupsen/logrus"
)

var version = "dev" // set by build process

var (
	listenAddr  = flag.String("listen", "[::1]:784", "address to listen on")
	backend     = flag.String("backend", "[::1]:53", "address of backend (UDP) DNS server")
	tlsCert     = flag.String("tlsCert", "cert.pem", "TLS certificate file")
	tlsKey      = flag.String("tlsKey", "key.pem", "TLS key file")
	tlsCompat   = flag.Bool("tlsCompat", false, "enable TLS compatibility mode")
	maxProcs    = flag.Int("maxProcs", 1, "GOMAXPROCS")
	verbose     = flag.Bool("verbose", false, "enable debug logging")
	showVersion = flag.Bool("version", false, "show version")
)

func main() {
	flag.Parse()

	// Evaluate flags
	if *verbose {
		log.SetLevel(log.DebugLevel)
		log.Debugln("enabled debug logging")
	}

	if *showVersion {
		log.Printf("doq https://github.com/natesales/doq version %s\n", version)
		os.Exit(1)
	}

	log.Debugf("tlsCompat: %v, tlsCert: %s, tlsKey: %s", *tlsCompat, *tlsCert, *tlsKey)

	// Set runtime GOMAXPROCS limit to limit goroutine system resource exhaustion
	runtime.GOMAXPROCS(*maxProcs)

	// Parse the TLS x509 keypair
	cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
	if err != nil {
		log.Fatalf("load TLS x509 cert: %s\n", err)
	}

	// Create the QUIC listener
	doqServer, err := server.New(*listenAddr, cert, *backend, *tlsCompat)
	if err != nil {
		log.Fatal(err)
	}

	// Accept QUIC connections
	log.Infof("starting quic listener on quic://%s\n", *listenAddr)
	doqServer.Listen()
}
