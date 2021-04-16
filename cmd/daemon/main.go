package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Compat   bool `short:"z" long:"compat" description:"Enable TLS backwards compatibility mode"`
	Insecure bool `short:"i" long:"insecure" description:"Ignore TLS certificate validation errors"`
	Verbose  bool `short:"v" long:"verbose" description:"Enable verbose logging"`
}

var options Options

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
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
