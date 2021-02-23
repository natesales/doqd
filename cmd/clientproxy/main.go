package main

import (
	"flag"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/doq/pkg/client"
)

var (
	doqServer             = flag.String("server", "[::1]:784", "DoQ server")
	tlsInsecureSkipVerify = flag.Bool("insecureSkipVerify", false, "skip TLS certificate validation")
	tlsCompat             = flag.Bool("tlsCompat", false, "enable TLS compatibility mode")
	listenAddr            = flag.String("listen", "[::1]:5353", "udp listen address")
)

var doqClient client.Client

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		resp, err := doqClient.SendQuery(*m)
		if err != nil {
			w.WriteMsg(&resp)
		}
	}

	w.WriteMsg(&dns.Msg{})
}

func main() {
	flag.Parse()

	// Create a new DoQ client
	var err error
	doqClient, err = client.New(*doqServer, *tlsInsecureSkipVerify, *tlsCompat)
	if err != nil {
		log.Fatal(doqClient)
	}
	defer doqClient.Close()

	dns.HandleFunc(".", handleDnsRequest)

	server := &dns.Server{Addr: *listenAddr, Net: "udp"}
	log.Infof("starting DNS server on %s\n", *listenAddr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
	defer server.Shutdown()
}
