package main

import (
	"crypto/tls"
	"os"
	"os/signal"
	"syscall"

	"github.com/natesales/doqd/pkg/server"
	log "github.com/sirupsen/logrus"
)

type ServerCommand struct {
	Listen      []string `short:"l" long:"listen" description:"Address to listen on" required:"true"`
	MetricsAddr string   `short:"m" long:"metrics" description:"Prometheus meterics listen address" required:"false"`
	Upstream    string   `short:"u" long:"upstream" description:"Upstream DNS server" required:"true"`
	Cert        string   `short:"c" long:"cert" description:"TLS certificate file" required:"true"`
	Key         string   `short:"k" long:"key" description:"TLS private key file" required:"true"`
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

	// Start metrics server
	go func() {
		log.Infof("Starting metrics server on %s", s.MetricsAddr)
		log.Fatal(server.MetricsListen(s.MetricsAddr))
	}()

	log.Debugf("Listening on %+v", s.Listen)
	for _, listenAddr := range s.Listen {
		// Create the QUIC listener
		doqServer, err := server.New(listenAddr, cert, s.Upstream, options.Compat)
		if err != nil {
			return nil
		}

		// Accept QUIC connections
		log.Infof("Starting QUIC listener on %s\n", listenAddr)
		go doqServer.Listen()
	}

	// Block until interrupt
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	return nil
}
