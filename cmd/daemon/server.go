package main

import (
	"crypto/tls"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/doq/pkg/server"
)

type ServerCommand struct {
	Listen   string `short:"l" long:"listen" description:"Address to listen on" required:"true" default:":8853"`
	Upstream string `short:"u" long:"upstream" description:"Upstream DNS server" required:"true"`
	Cert     string `short:"c" long:"cert" description:"TLS certificate file" required:"true"`
	Key      string `short:"k" long:"key" description:"TLS private key file" required:"true"`
}

var serverCommand ServerCommand

func init() {
	if _, err := parser.AddCommand(
		"server",
		"DoQ server proxy",
		"Start a DoQ server proxy",
		&serverCommand); err != nil {
		log.Fatal(err)
	}
}

func (s *ServerCommand) Execute(args []string) error {
	// Load the keypair for TLS
	cert, err := tls.LoadX509KeyPair(s.Cert, s.Key)
	if err != nil {
		log.Fatalf("load TLS x509 cert: %s\n", err)
	}

	// Create the QUIC listener
	doqServer, err := server.New(s.Listen, cert, s.Upstream, options.Compat)
	if err != nil {
		return nil
	}

	// Accept QUIC connections
	log.Infof("starting QUIC listener on %s\n", s.Listen)
	doqServer.Listen()

	return nil
}
