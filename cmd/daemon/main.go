package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"io/ioutil"
	"net"
	"os"
	"runtime"

	"github.com/lucas-clemente/quic-go"
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
	listener, err := quic.ListenAddr(*listenAddr, &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"dq"},
	}, nil)
	if err != nil {
		log.Fatalf("quic listen: %s\n", err)
	}
	defer listener.Close() // Clean up the listener once we're done with it

	// Accept QUIC connections
	for {
		session, err := listener.Accept(context.Background())
		if err != nil {
			log.Fatalf("quic listen accept: %s\n", err)
		}

		// Handle the client connection in a new goroutine
		go func() {
			err := handleClient(session)
			if err != nil {
				log.Warn(err)
			}
		}()
	}
}

// handleClient handles a DoQ quic.Session
func handleClient(session quic.Session) error {
	for {
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			return err
		}

		streamErrChan := make(chan error)

		// Handle each QUIC stream in a new goroutine
		go func() {
			defer stream.Close() // Clean up the stream once we're done with it

			data, err := ioutil.ReadAll(stream)
			if err != nil {
				log.Warnf("read query: %v", err)
			}

			// Connect to the DNS backend
			conn, err := net.Dial("udp", *backend)
			if err != nil {
				streamErrChan <- errors.New("backend connect: " + err.Error())
				return
			}

			// Send query to DNS backend
			_, err = conn.Write(data)
			if err != nil {
				streamErrChan <- errors.New("backend query write: " + err.Error())
				return
			}

			// Read the query response from the backend
			buf := make([]byte, 4096)
			size, err := conn.Read(buf)
			if err != nil {
				streamErrChan <- errors.New("backend query read: " + err.Error())
				return
			}
			buf = buf[:size]

			// Write the response to the QUIC stream
			_, err = stream.Write(buf)
			if err != nil {
				streamErrChan <- errors.New("quic stream write: " + err.Error())
				return
			}
		}()

		// Retrieve the stream error
		err = <-streamErrChan
		if err != nil {
			log.Warn(err)
			break // Close the connection
		}
	}

	// Close the QUIC session
	_ = session.CloseWithError(0, "") // Ignore error - if we're already closing the session it doesn't matter if it errors or not

	return nil // nil error
}
