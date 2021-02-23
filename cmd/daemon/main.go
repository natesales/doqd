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
	listenAddr  = flag.String("bind", "[::1]:784", "address to listen on")
	backend     = flag.String("backend", "[::1]:53", "address of backend (UDP) DNS server")
	tlsCert     = flag.String("tlsCert", "cert.pem", "TLS certificate file")
	tlsKey      = flag.String("tlsKey", "key.pem", "TLS key file")
	maxProcs    = flag.Int("maxProcs", 1, "GOMAXPROCS")
	showVersion = flag.Bool("version", false, "show version")
)

func main() {
	flag.Parse()

	if *showVersion {
		log.Printf("doq https://github.com/natesales/doq version %s\n", version)
		os.Exit(1)
	}

	// Set runtime GOMAXPROCS limit to limit goroutine system resource exhaustion
	runtime.GOMAXPROCS(*maxProcs)

	// Parse the TLS x509 keypair
	cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
	if err != nil {
		log.Fatalf("load TLS x509 cert: %s\n", err)
	}

	// Create the QUIC listener
	log.Infof("starting quic listener on %s\n", *listenAddr)
	doqServer, err := server.New(*listenAddr, cert, *backend)
	if err != nil {
		log.Fatal(err)
	}
	defer doqServer.Close() // Clean up the listener once we're done with it

	// Accept QUIC connections
	err = doqServer.Listen()
	if err != nil {
		log.Fatal(err)
	}
}
