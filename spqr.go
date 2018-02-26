/*
 * Copyright (c) 2018, Jeremy Bingham (<jeremy@goiardi.gl>)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/* A consul based user management tool. */

package main

import (
	"encoding/json"
	"github.com/ctdk/spqr/internal/config"
	"github.com/ctdk/spqr/internal/state"
	consul "github.com/hashicorp/consul/api"
	"github.com/tideland/golib/logger"
	"os"
)

func main() {
	config.ParseConfigOptions()

	consulClient, err := configureConsul()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	logger.Debugf("connected to consul")

	var stateHolder *state.State
	inCh := make(chan *state.Indices)
	errCh := make(chan error)
	doneCh := make(chan struct{})

	if config.Config.StateFile != "" {
		logger.Debugf("setting up the state file")
		go state.InitState(&stateHolder, config.Config.StateFile, inCh, errCh, doneCh)
		err = <-errCh
		if err != nil {
			logger.Fatalf(err.Error())
		}
	} else {
		logger.Debugf("no state file configured")
	}

	// JSON incoming!
	var incoming interface{}
	dec := json.NewDecoder(os.Stdin)
	dec.UseNumber()

	if err := dec.Decode(&incoming); err != nil {
		logger.Errorf(err.Error())
	}

	logger.Debugf("incoming: %T %v", incoming, incoming)

	switch incoming := incoming.(type) {
	case nil:
		logger.Debugf("nil event, won't do anything")
		var h *state.Indices
		inCh <- h
		close(inCh)
	case []interface{}:
		if len(incoming) == 0 {
			logger.Debugf("empty item, skipping")
			break
		}
		logger.Debugf("key prefix or event, probably (don't care about the other possibilities)")
		handleIncoming(consulClient, stateHolder, inCh, incoming)
	default:
		logger.Debugf("Not anything we're interested in: %T", incoming)
		var h *state.Indices
		inCh <- h
		close(inCh)
	}
	<-doneCh
	logger.Debugf("received done signal, exiting")
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
