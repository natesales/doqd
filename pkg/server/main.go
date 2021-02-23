package server

import (
	"context"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"time"

	"github.com/lucas-clemente/quic-go"
	log "github.com/sirupsen/logrus"
)

// Config constants
var (
	QuicProtos         = []string{"doq-i02", "doq-i00", "dq", "doq"}
	QuicMaxIdleTimeout = 5 * time.Minute
	DnsMinPacketSize   = 12 + 5
)

// Protocol constants
const (
	DoqNoError       = 0x00 // No error. This is used when the connection or stream needs to be closed, but there is no error to signal.
	DoqInternalError = 0x01 // The DoQ implementation encountered an internal error and is incapable of pursuing the transaction or the connection
)

// Server stores a DoQ server
type Server struct {
	Backend  string
	Listener quic.Listener
}

// New constructs a new Server
func New(addr string, cert tls.Certificate, backend string) (Server, error) {
	// Create the QUIC listener
	listener, err := quic.ListenAddr(addr, &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   QuicProtos,
	}, &quic.Config{MaxIdleTimeout: QuicMaxIdleTimeout})
	if err != nil {
		return Server{}, errors.New("quic listen: " + err.Error())
	}

	return Server{
		Backend:  backend,
		Listener: listener,
	}, nil // nil error
}

// Close closes the server QUIC listener
func (s Server) Close() error {
	return s.Listener.Close()
}

// Listen starts the QUIC listener
func (s Server) Listen() error {
	// Accept QUIC connections
	for {
		session, err := s.Listener.Accept(context.Background())
		if err != nil {
			return errors.New("quic listen accept: " + err.Error())
		}

		// Handle the client connection in a new goroutine
		go func() {
			err := handleClient(session, s.Backend)
			if err != nil {
				log.Warn(err)
			}
		}()
	}
}

// handleClient handles a DoQ quic.Session
func handleClient(session quic.Session, backend string) error {
	for {
		stream, err := session.AcceptStream(context.Background())
		if err != nil { // Close the session if we aren't able to accept the incoming stream
			_ = session.CloseWithError(DoqInternalError, "")
			return nil
		}

		streamErrChannel := make(chan error)

		// Handle each QUIC stream in a new goroutine
		go func() {
			defer stream.Close() // Clean up the stream once we're done with it

			// Read the DNS message
			data, err := ioutil.ReadAll(stream)
			if err != nil {
				streamErrChannel <- errors.New("read query: " + err.Error())
				return
			}
			if len(data) < DnsMinPacketSize {
				streamErrChannel <- errors.New("dns packet too small")
				return
			}

			// Connect to the DNS backend
			conn, err := net.Dial("udp", backend)
			if err != nil {
				streamErrChannel <- errors.New("backend connect: " + err.Error())
				return
			}

			// Send query to DNS backend
			_, err = conn.Write(data)
			if err != nil {
				streamErrChannel <- errors.New("backend query write: " + err.Error())
				return
			}

			// Read the query response from the backend
			buf := make([]byte, 4096)
			size, err := conn.Read(buf)
			if err != nil {
				streamErrChannel <- errors.New("backend query read: " + err.Error())
				return
			}
			buf = buf[:size]

			// Write the response to the QUIC stream
			_, err = stream.Write(buf)
			if err != nil {
				streamErrChannel <- errors.New("quic stream write: " + err.Error())
				return
			}
		}()

		// Retrieve the stream error
		err = <-streamErrChannel
		if err != nil {
			log.Warn(err)
			break // Close the connection
		}
	}

	// Close the QUIC session
	_ = session.CloseWithError(DoqNoError, "") // Ignore error - if we're already closing the session it doesn't matter if it errors or not

	return nil // nil error
}
