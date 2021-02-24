package main

import (
	"flag"
	"net"
	"os"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/doq/pkg/client"
)

var (
	doqServer             = flag.String("server", "[::1]:784", "DoQ server")
	tlsInsecureSkipVerify = flag.Bool("insecureSkipVerify", false, "skip TLS certificate validation")
	tlsCompat             = flag.Bool("tlsCompat", false, "enable TLS compatibility mode")
	listenAddr            = flag.String("listen", "[::1]:5353", "udp listen address")
	verbose               = flag.Bool("verbose", false, "enable debug logging")
)

func main() {
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	// Create the UDP DNS listener
	log.Debug("creating UDP listener")
	pc, err := net.ListenPacket("udp", *listenAddr)
	if err != nil {
		return
	}
	defer pc.Close()

	log.Debugln("ready to accept connections")
	for {
		log.Debugln("ready to read from buffer")
		buffer := make([]byte, 4096)
		n, addr, err := pc.ReadFrom(buffer)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		log.Debugf("read %d bytes from buffer", n)

		// Unpack the DNS message
		log.Debugln("unpacking dns message")
		var msgIn dns.Msg
		err = msgIn.Unpack(buffer)
		if err != nil {
			log.Warn(err)
		}

		// Create a new DoQ client
		// TODO: Fix this, the client should be able to send multiple queries on a connection
		doqClient, err := client.New(*doqServer, *tlsInsecureSkipVerify, *tlsCompat)
		if err != nil {
			log.Fatal(doqClient)
			os.Exit(1)
		}

		// Send the DoQ query
		log.Debugln("sending DoQ query")
		resp, err := doqClient.SendQuery(msgIn)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		log.Debugln("closing doq quic stream")
		doqClient.Close()

		// Pack the response DNS message to wire format
		log.Debugln("packing response dns message")
		packed, err := resp.Pack()
		if err != nil {
			log.Fatal(err)
		}

		// Write response to UDP connection
		log.Debugln("writing response dns message")
		_, err = pc.WriteTo(packed, addr)
		if err != nil {
			log.Fatal(err)
		}
		log.Debug("finished writing")
	}
}
