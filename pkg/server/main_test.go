package server

import (
	"crypto/tls"
	"testing"

	"github.com/miekg/dns"

	"github.com/natesales/doq/pkg/client"
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

	doqClient, err := client.New("localhost:8853", true, false)
	if err != nil {
		t.Error(err)
	}

	req := dns.Msg{
		Question: []dns.Question{{
			Name:   dns.Fqdn("example.com"),
			Qtype:  dns.StringToType["A"],
			Qclass: dns.ClassINET,
		}},
	}
	req.RecursionDesired = true

	_, err = doqClient.SendQuery(req)
	if err != nil {
		t.Error(err)
	}
}
