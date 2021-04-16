package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

var version = "dev" // set by build process

type Options struct {
	Compat      bool `short:"z" long:"compat" description:"Enable TLS backwards compatibility mode"`
	Insecure    bool `short:"i" long:"insecure" description:"Ignore TLS certificate validation errors"`
	Verbose     bool `short:"v" long:"verbose" description:"Enable verbose logging"`
	ShowVersion bool `short:"V" long:"version" description:"Show version and exit"`
}

var options Options

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		// Enable debug logging in development releases
		if options.Verbose {
			log.SetLevel(log.DebugLevel)
		}

		if options.ShowVersion {
			log.Printf("doq version %s https://github.com/natesales/doqd", version)
			os.Exit(0)
		}

		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}
