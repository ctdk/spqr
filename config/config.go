package config

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"runtime"
)

const Version = "0.0.1"

var GitHash = "unknown"

type Conf struct {
	ConsulHttpAddr string
}

type Options struct {
	Version bool `short:"v" long:"version" description:"Print version info."`
	ConsulHttpAddr string `short:"C" long:"consul-http-addr" description:"Consul HTTP API endpoint. Defaults to http://127.0.0.1:8500. Shares the same CONSUL_HTTP_ADDR environment variable as consul itself as well." env:"CONSUL_HTTP_ADDR"`
}

func initConfig() *Conf { return &Conf{} }

var Config = initConfig()

func ParseConfigOptions() error {
	var opts = &Options{}
	parser := flags.NewParser(opts, flags.Default)
	parser.ShortDescription = fmt.Sprintf("A consul leveraging user account manager - version %s", Version)

	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		if err.(*flags.Error).Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Println(err)
			os.Exit(1)
		}
	}
	if opts.Version {
		fmt.Printf("spqr version %s (git hash: %s) built with %s.\n", Version, GitHash, runtime.Version())
		os.Exit(0)
	}
	if opts.ConsulHttpAddr != "" {
		Config.ConsulHttpAddr = opts.ConsulHttpAddr
	}
	if Config.ConsulHttpAddr == "" {
		Config.ConsulHttpAddr = "http://127.0.0.1:8500"
	}

	return nil
}
