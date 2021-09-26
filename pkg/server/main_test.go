package server

import (
	"crypto/tls"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"

	"github.com/natesales/doqd/pkg/client"
)

func TestServer(t *testing.T) {
	// Load the keypair for TLS
	cert, err := tls.LoadX509KeyPair("/tmp/cert.pem", "/tmp/key.pem")
	assert.Nil(t, err)

	// Create the QUIC listener
	doqServer, err := New("localhost:8853", cert, "1.1.1.1:53", false)
	assert.Nil(t, err)

	// Start the server
	go doqServer.Listen()

	// Create the DoQ client
	doqClient, err := client.New("localhost:8853", true, false)
	assert.Nil(t, err)

	// Create a test DNS query
	req := dns.Msg{
		Question: []dns.Question{{
			Name:   dns.Fqdn("example.com"),
			Qtype:  dns.StringToType["A"],
			Qclass: dns.ClassINET,
		}},
	}
	req.RecursionDesired = true

	// Send the query
	_, err = doqClient.SendQuery(req)
	assert.Nil(t, err)
}
