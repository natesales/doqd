package main

import (
	"flag"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"net"
	"os"

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

	// Create a new DoQ client
	doqClient, err := client.New(*doqServer, *tlsInsecureSkipVerify, *tlsCompat)
	if err != nil {
		log.Fatal(doqClient)
		os.Exit(1)
	}
	defer doqClient.Close()

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

		var msgIn dns.Msg
		err = msgIn.Unpack(buffer)
		if err != nil {
			log.Warn(err)
		}

		log.Debugln("sending DoQ query")
		resp, err := doqClient.SendQuery(msgIn)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		packed, err := resp.Pack()
		pc.WriteTo(packed, addr)
		log.Debug("finished writing")
	}
}
