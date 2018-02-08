package main

import (
	"encoding/json"
	"fmt"
	"github.com/ctdk/spqr/config"
	_ "github.com/ctdk/spqr/users"
	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	"log"
	"os"
)

// TODO: provide more configuration options
func main() {
	config.ParseConfigOptions()

	// configure vault
	vaultClient, err := configureVault()
	if err != nil {
		log.Fatal(err)
	}
	consulClient, err := configureConsul()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to consul")
	_ = vaultClient


	// JSON incoming!
	var incoming interface{}
	dec := json.NewDecoder(os.Stdin)
	dec.UseNumber()

	if err := dec.Decode(&incoming); err != nil {
		log.Println(err)
	}

	fmt.Printf("incoming: %T %v\n", incoming, incoming)

	switch incoming := incoming.(type) {
	case nil:
		fmt.Printf("won't do anything\n")
	case []interface{}:
		if len(incoming) == 0 {
			log.Println("empty item, skipping")
			break
		}
		fmt.Printf("key prefix or event, probably (don't care about the other possibilities)\n")
		handleIncoming(consulClient, vaultClient, incoming)
	default:
		fmt.Printf("Not anything we're interested in: %T\n", incoming)
	}
	

	// moo.
	/*
	_, err = users.New("foomer", "Foo Mer", "", "", []string{}, []string{"badkey"})
	if err != nil {
		log.Fatal(err)
	}
	u, err := users.Get("foomer")
	if err != nil {
		log.Fatal(err)
	}
	u.SSHKeys = []string{"aBadKey","soBad","aVeryBadKeyIndeed","wayBad"}
	err = u.Update()
	if err != nil {
		log.Fatal(err)
	}
	err = u.Disable()
	if err != nil {
		log.Fatal(err)
	}
	*/
}

func configureVault() (*vault.Client, error) {
	conf := vault.DefaultConfig()
	if err := conf.ReadEnvironment(); err != nil {
		return nil, err
	}
	c, err := vault.NewClient(conf)
	if err != nil {
		return nil, err
	}

	return c, nil
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
