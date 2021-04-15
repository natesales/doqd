package main

import (
	"crypto/tls"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/doq/pkg/server"
)

//goland:noinspection GoUnusedGlobalVariable
var version = "dev" // set by build process

// CLI Flags
//goland:noinspection GoUnusedGlobalVariable
var opts struct {
	Listen   string `short:"l" long:"listen" description:"Address to listen on" required:"true" default:":8853"`
	Upstream string `short:"u" long:"upstream" description:"Upstream DNS server" required:"true"`
	Cert     string `short:"c" long:"cert" description:"TLS certificate file" required:"true"`
	Key      string `short:"k" long:"key" description:"TLS private key file" required:"true"`
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

	// Load the keypair for TLS
	cert, err := tls.LoadX509KeyPair(opts.Cert, opts.Key)
	if err != nil {
		log.Fatalf("load TLS x509 cert: %s\n", err)
	}

	// Create the QUIC listener
	doqServer, err := server.New(opts.Listen, cert, opts.Upstream, opts.Compat)
	if err != nil {
		log.Fatal(err)
	}

	// Accept QUIC connections
	log.Infof("starting QUIC listener on %s\n", opts.Listen)
	doqServer.Listen()
}
