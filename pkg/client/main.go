package client

import (
	"crypto/tls"
	"errors"
	"io/ioutil"

	"github.com/lucas-clemente/quic-go"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

// Client stores a DoQ client
type Client struct {
	Session quic.Session
}

// New constructs a new client
func New(server string, tlsInsecureSkipVerify bool) (Client, error) {
	// Connect to DoQ server
	session, err := quic.DialAddr(server, &tls.Config{
		InsecureSkipVerify: tlsInsecureSkipVerify,
		NextProtos:         []string{"dq"},
	}, nil)
	if err != nil {
		log.Fatal("failed to connect to the server: %v\n", err)
	}

	return Client{session}, nil // nil error
}

// Close closes a Client QUIC connection
func (c Client) Close() error {
	return c.Session.CloseWithError(0, "")
}

// SendQuery sends query over a new QUIC stream
func (c Client) SendQuery(message dns.Msg) (dns.Msg, error) {
	// Open a new QUIC stream
	stream, err := c.Session.OpenStream()
	if err != nil {
		return dns.Msg{}, errors.New("quic stream open: " + err.Error())
	}

	// Pack the DNS message for transmission
	wire, err := message.Pack()
	if err != nil {
		stream.Close()
		return dns.Msg{}, errors.New("dns message pack: " + err.Error())
	}

	// Send the DNS query over QUIC
	_, err = stream.Write(wire)
	stream.Close()
	if err != nil {
		return dns.Msg{}, errors.New("quic stream write: " + err.Error())
	}

	// Read the response
	response, err := ioutil.ReadAll(stream)
	if err != nil {
		return dns.Msg{}, errors.New("quic stream read: " + err.Error())
	}

	// Unpack the DNS message
	var msg dns.Msg
	err = msg.Unpack(response)
	if err != nil {
		return dns.Msg{}, errors.New("dns message unpack: " + err.Error())
	}

	return msg, nil // nil error
}
