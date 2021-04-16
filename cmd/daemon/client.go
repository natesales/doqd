package main

import (
	"net"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/doq/pkg/client"
)

type ClientCommand struct {
	Listen   string `short:"l" long:"listen" description:"Address to listen on" required:"true" default:":53"`
	Upstream string `short:"u" long:"upstream" description:"Upstream DNS server" required:"true" default:":8853"`
}

var clientCommand ClientCommand

func init() {
	if _, err := parser.AddCommand(
		"client",
		"DoQ client proxy",
		"Start a DoQ client proxy",
		&clientCommand); err != nil {
		log.Fatal(err)
	}
}

func (c *ClientCommand) Execute(args []string) error {
	// Create the UDP DNS listener
	log.Infof("starting UDP listener on %s\n", c.Listen)
	pc, err := net.ListenPacket("udp", c.Listen)
	if err != nil {
		log.Fatal(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer pc.Close()

	log.Debugln("ready to accept connections")
	for {
		log.Debugln("ready to read from buffer")
		buffer := make([]byte, 4096)
		n, addr, err := pc.ReadFrom(buffer)
		if err != nil {
			log.Warn(err)
		}
		log.Debugf("read %d bytes from buffer", n)

		// Unpack the DNS message
		log.Debugln("unpacking DNS message")
		var msgIn dns.Msg
		err = msgIn.Unpack(buffer)
		if err != nil {
			log.Warn(err)
		}

		// Create a new DoQ client
		log.Debugf("opening QUIC connection to %s\n", c.Upstream)
		doqClient, err := client.New(c.Upstream, options.Insecure, options.Compat)
		if err != nil {
			log.Warn(err)
		}

		// Send the DoQ query
		log.Debugln("sending DoQ query")
		resp, err := doqClient.SendQuery(msgIn)
		if err != nil {
			log.Warn(err)
		}
		log.Debugln("closing doq QUIC stream")
		_ = doqClient.Close()

		// Pack the response DNS message to wire format
		log.Debugln("packing response DNS message")
		packed, err := resp.Pack()
		if err != nil {
			log.Warn(err)
		}

		// Write response to UDP connection
		log.Debugln("writing response DNS message")
		_, err = pc.WriteTo(packed, addr)
		if err != nil {
			log.Warn(err)
		}
		log.Debug("finished writing")
	}
}
