package server

import (
	"crypto/tls"
	"testing"

	"github.com/miekg/dns"

	"github.com/natesales/doqd/pkg/client"
)

func TestServer(t *testing.T) {
	// Load the keypair for TLS
	cert, err := tls.LoadX509KeyPair("/tmp/cert.pem", "/tmp/key.pem")
	if err != nil {
		t.Error(err)
	}

	// Create the QUIC listener
	doqServer, err := New("localhost:8853", cert, "1.1.1.1:53", false)
	if err != nil {
		t.Error(err)
	}

	// Start the server
	go doqServer.Listen()

	// Create the DoQ client
	doqClient, err := client.New("localhost:8853", true, false)
	if err != nil {
		t.Error(err)
	}

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
	if err != nil {
		t.Error(err)
	}
}
