package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/lucas-clemente/quic-go"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

var (
	server    = flag.String("server", "[::1]:784", "DoQ server")
	dnssec    = flag.Bool("dnssec", true, "send DNSSEC flag")
	rec       = flag.Bool("recursion", true, "send RD flag")
	queryName = flag.String("queryName", "", "DNS QNAME")
	queryType = flag.String("queryType", "", "DNS QTYPE")
)

func main() {
	flag.Parse()

	// Validate flags
	if *queryName == "" || *queryType == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse QNAME/QTYPE
	qname := dns.Fqdn(*queryName)
	qtype, success := dns.StringToType[*queryType]
	if !success {
		log.Fatalf("invalid DNS QTYPE \"%s\"\n", *queryType)
	}

	// Connect to DoQ server
	session, err := quic.DialAddr(*server, &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"dq"},
	}, nil)
	if err != nil {
		log.Fatal("failed to connect to the server: %v\n", err)
	}
	defer session.CloseWithError(0, "") // Cleanup the QUIC session once we're done with it

	// Open a new QUIC stream
	stream, err := session.OpenStream()
	if err != nil {
		log.Fatalf("quic stream open: %s\n", err)
	}

	// Create the DNS query message
	msg := dns.Msg{}
	msg.SetQuestion(qname, qtype)
	msg.SetEdns0(4096, *dnssec)
	msg.RecursionDesired = *rec
	wire, err := msg.Pack()
	if err != nil {
		stream.Close()
		log.Fatalf("dns message pack: %s\n", err)
	}

	// Send the DNS query over QUIC
	_, err = stream.Write(wire)
	stream.Close()
	if err != nil {
		log.Fatalf("quic stream write: %s\n", err)
	}

	// Read the response
	response, err := ioutil.ReadAll(stream)
	if err != nil {
		log.Fatalf("quic stream read: %s\n", err)
	}

	// Unpack the DNS message
	err = msg.Unpack(response)
	if err != nil {
		log.Fatalf("dns message unpack: %s\n", err)
	}

	fmt.Println(&msg)
}
