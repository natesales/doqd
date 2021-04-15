package main

import (
	"net"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/doq/pkg/client"
)

//goland:noinspection GoUnusedGlobalVariable
var version = "dev" // set by build process

// CLI Flags
//goland:noinspection GoUnusedGlobalVariable
var opts struct {
	Listen   string `short:"l" long:"listen" description:"Address to listen on" required:"true" default:":53"`
	Upstream string `short:"u" long:"upstream" description:"Upstream DNS server" required:"true" default:":8853"`
	Insecure bool   `short:"i" long:"insecure" description:"Ignore TLS certificate validation errors"`
	Compat   bool   `short:"z" long:"compat" description:"Enable TLS backwards compatibility mode"`
	Verbose  bool   `short:"v" long:"verbose" description:"Enable verbose logging"`
}

func main() {
	// Parse cli flags
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		if !strings.Contains(err.Error(), "Usage") {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	// Enable debug logging in development releases
	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// Create the UDP DNS listener
	log.Infof("starting UDP listener on %s\n", opts.Listen)
	pc, err := net.ListenPacket("udp", opts.Listen)
	if err != nil {
		log.Fatal(err)
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
		log.Debugf("opening QUIC connection to %s\n", opts.Upstream)
		doqClient, err := client.New(opts.Upstream, opts.Insecure, opts.Compat)
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
		_ = doqClient.Close()

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
