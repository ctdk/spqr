package main

import (
	"encoding/json"
	"fmt"
	"github.com/ctdk/spqr/config"
	consul "github.com/hashicorp/consul/api"
	"github.com/tideland/golib/logger"
	"os"
)

// TODO: provide more configuration options
func main() {
	config.ParseConfigOptions()

	consulClient, err := configureConsul()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	logger.Debugf("connected to consul")

	// JSON incoming!
	var incoming interface{}
	dec := json.NewDecoder(os.Stdin)
	dec.UseNumber()

	if err := dec.Decode(&incoming); err != nil {
		logger.Errorf(err.Error())
	}

	logger.Debugf("incoming: %T %v\n", incoming, incoming)

	switch incoming := incoming.(type) {
	case nil:
		fmt.Printf("won't do anything\n")
	case []interface{}:
		if len(incoming) == 0 {
			logger.Debugf("empty item, skipping")
			break
		}
		fmt.Printf("key prefix or event, probably (don't care about the other possibilities)\n")
		handleIncoming(consulClient, incoming)
	default:
		logger.Debugf("Not anything we're interested in: %T\n", incoming)
	}
}

func configureConsul() (*consul.Client, error) {
	conf := consul.DefaultConfig()

	conf.Address = config.Config.ConsulHttpAddr

	c, err := consul.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return c, nil
}
