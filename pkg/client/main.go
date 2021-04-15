package client

import (
	"crypto/tls"
	"errors"
	"io/ioutil"

	"github.com/lucas-clemente/quic-go"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/doq"
)

// Client stores a DoQ client
type Client struct {
	Session quic.Session
}

// New constructs a new client
func New(server string, tlsInsecureSkipVerify bool, compat bool) (Client, error) {
	// Select TLS protocols for DoQ
	var tlsProtos []string
	if compat {
		tlsProtos = doq.TlsProtosCompat
	} else {
		tlsProtos = doq.TlsProtos
	}

	// Connect to DoQ server
	log.Debugln("dialing quic server")
	session, err := quic.DialAddr(server, &tls.Config{
		InsecureSkipVerify: tlsInsecureSkipVerify,
		NextProtos:         tlsProtos,
	}, nil)
	if err != nil {
		log.Fatalf("failed to connect to the server: %v\n", err)
	}

	return Client{session}, nil // nil error
}

// Close closes a Client QUIC connection
func (c Client) Close() error {
	log.Debugln("closing quic session")
	return c.Session.CloseWithError(0, "")
}

// SendQuery sends query over a new QUIC stream
func (c Client) SendQuery(message dns.Msg) (dns.Msg, error) {
	// Open a new QUIC stream
	log.Debugln("opening new quic stream")
	stream, err := c.Session.OpenStream()
	if err != nil {
		return dns.Msg{}, errors.New("quic stream open: " + err.Error())
	}

	// Pack the DNS message for transmission
	log.Debugln("packing dns message")
	packed, err := message.Pack()
	if err != nil {
		_ = stream.Close()
		return dns.Msg{}, errors.New("dns message pack: " + err.Error())
	}

	// Send the DNS query over QUIC
	log.Debugln("writing packed format to the stream")
	_, err = stream.Write(packed)
	_ = stream.Close()
	if err != nil {
		return dns.Msg{}, errors.New("quic stream write: " + err.Error())
	}

	// Read the response
	log.Debugln("reading server response")
	response, err := ioutil.ReadAll(stream)
	if err != nil {
		return dns.Msg{}, errors.New("quic stream read: " + err.Error())
	}

	// Unpack the DNS message
	log.Debugln("unpacking response dns message")
	var msg dns.Msg
	err = msg.Unpack(response)
	if err != nil {
		return dns.Msg{}, errors.New("dns message unpack: " + err.Error())
	}

	return msg, nil // nil error
}
