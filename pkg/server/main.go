package server

import (
	"context"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/doq"
)

// Server stores a DoQ server
type Server struct {
	Upstream string
	Listener quic.Listener
}

// New constructs a new Server
func New(listenAddr string, cert tls.Certificate, upstream string, tlsCompat bool) (*Server, error) {
	// Select TLS protocols for DoQ
	var tlsProtos []string
	if tlsCompat {
		tlsProtos = doq.TlsProtosCompat
	} else {
		tlsProtos = doq.TlsProtos
	}

	// Create QUIC listener
	listener, err := quic.ListenAddr(listenAddr, &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   tlsProtos,
	}, &quic.Config{MaxIdleTimeout: 5 * time.Second})
	if err != nil {
		return nil, errors.New("could not start QUIC listener: " + err.Error())
	}

	return &Server{Listener: listener, Upstream: upstream}, nil // nil error
}

// Listen starts accepting QUIC connections
func (s *Server) Listen() {
	// Accept QUIC connections
	for {
		session, err := s.Listener.Accept(context.Background())
		if err != nil {
			log.Infof("QUIC accept: %v", err)
			break
		} else {
			// Handle QUIC session in a new goroutine
			go handleDoQSession(session, s.Upstream)
		}
	}
}

// handleDoQSession handles a new DoQ session
func handleDoQSession(session quic.Session, upstream string) {
	for {
		// Accept client-originated QUIC stream
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			log.Warnf("QUIC stream accept: %v", err)
			_ = session.CloseWithError(doq.InternalError, "") // Close the session with an internal error message
			return
		}

		// Handle QUIC stream (DNS query) in a new goroutine
		go func() {
			// The client MUST send the DNS query over the selected stream, and MUST
			// indicate through the STREAM FIN mechanism that no further data will
			// be sent on that stream.
			bytes, err := ioutil.ReadAll(stream) // Ignore error, error handling is done by packet length

			// Check for packet to small
			if len(bytes) < 17 { // MinDnsPacketSize
				switch {
				case err != nil:
					log.Infof("QUIC stream read: %v", err)
				default:
					log.Info("DNS query length is too small")
				}
				return
			}

			// Unpack the incoming DNS message
			msg := dns.Msg{}
			err = msg.Unpack(bytes)
			if err != nil {
				log.Infof("DNS query unpack: %v", err)
			}

			// If any message sent on a DoQ connection contains an edns-tcp-keepalive EDNS(0) Option,
			// this is a fatal error and the recipient of the defective message MUST forcibly abort
			// the connection immediately.
			if opt := msg.IsEdns0(); opt != nil {
				for _, option := range opt.Option {
					// Check for EDNS TCP keepalive option
					if option.Option() == dns.EDNS0TCPKEEPALIVE {
						_ = stream.Close() // Ignore error if we're already trying to forcibly close the stream
						return
					}
				}
			}

			// Query the upstream for our DNS response
			resp, err := sendUdpDnsMessage(msg, upstream)
			if err != nil {
				log.Warn(err)
			}

			// Pack the response into a byte slice
			bytes, err = resp.Pack()
			if err != nil {
				log.Warn(err)
			}

			// Send the byte slice over the open QUIC stream
			n, err := stream.Write(bytes)
			if err != nil {
				log.Warn(err)
			}
			if n != len(bytes) {
				log.Warn("stream write length mismatch")
			}

			// Ignore error since we're already trying to close the stream
			_ = stream.Close()
		}()
	}
}

func sendUdpDnsMessage(msg dns.Msg, upstream string) (dns.Msg, error) {
	// Pack the DNS message
	packed, err := msg.Pack()
	if err != nil {
		return dns.Msg{}, err
	}

	// Connect to the DNS upstream
	log.Debugln("dialing udp dns upstream")
	conn, err := net.Dial("udp", upstream)
	if err != nil {
		return dns.Msg{}, errors.New("upstream connect: " + err.Error())
	}

	// Send query to DNS upstream
	log.Debugln("writing query to dns upstream")
	_, err = conn.Write(packed)
	if err != nil {
		return dns.Msg{}, errors.New("upstream query write: " + err.Error())
	}

	// Read the query response from the upstream
	log.Debugln("reading query response")
	buf := make([]byte, 4096)
	size, err := conn.Read(buf)
	if err != nil {
		return dns.Msg{}, errors.New("upstream query read: " + err.Error())
	}

	// Pack the response message
	var retMsg dns.Msg
	err = retMsg.Unpack(buf[:size])
	if err != nil {
		return dns.Msg{}, err
	}

	return retMsg, nil // nil error
}
