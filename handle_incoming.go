package main

import (
	"github.com/ctdk/spqr/users"
	consul "github.com/hashicorp/consul/api"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
)

const (
	notAThing uint8 = iota
	keyPrefix
	consulEvent
)

var handleDesc = map[uint8]string{
	keyPrefix: "key prefix",
	consulEvent: "event",
}

func handleIncoming(c *consul.Client, keys []interface{}) {
	var handlingType uint8
	var groupLists [][]string

	for _, k := range keys {
		switch k := k.(type) {
		case map[string]interface{}:
			fmt.Printf("what I expected: %+v\n", k)
			var payload string
			
			// within one request, everything will be just one kind
			// of thing so this only needs to be checked once.
			if handlingType == notAThing {
				if r, ok := k["Value"].(string); ok {
					payload = r
					handlingType = keyPrefix
				} else if r, ok := k["Payload"].(string); ok {
					payload = r
					handlingType = consulEvent
				} else {
					log.Println("doesn't appear to be a keyprefix or event, moving on")
					continue
				}
			} else {
				var ok bool
				switch handlingType {
				case keyPrefix:
					payload, ok = k["Value"].(string)
				case consulEvent:
					payload, ok = k["Payload"].(string)
				default:
					panic("there's no way this should be reachable")
				}
				if !ok {
					log.Printf("expected a string, but something went wrong")
				}
			}

			val, err := base64.StdEncoding.DecodeString(payload)
			if err != nil {
				log.Println(err)
			}

			fmt.Printf("and val: '%s'\n", val)
			j := make(map[string]interface{})
			err = json.Unmarshal([]byte(val), &j)
			if err != nil {
				log.Println(err)
				continue
			}
			fmt.Printf("this is a %s\n", handleDesc[handlingType])
			for u, y := range j {
				fmt.Printf("u: %s y: %T %v\n", u, y, y)
			}
			switch handlingType {
			case keyPrefix:
				convUsers, err := convertUsersInterfaceSlice(j["users"].([]interface{}))
				if err != nil {
					log.Println(err)
					continue
				}
				groupLists = append(groupLists, convUsers)
			default:
				log.Printf("not handling %s yet in switch", handleDesc[handlingType])
			}
		default:
			fmt.Printf("NOT what I expected: %T %v", k, k)
		}
	}

	// So what do we do?
	switch handlingType {
	case keyPrefix:
		fmt.Printf("groups: %v\n", groupLists)
		if u2get, err := removeDupeUsers(groupLists); err != nil {
			log.Println(err)
		} else {
			fmt.Printf("cleaned up user list: %v\n", u2get)
			uc := users.NewUserConsulClient(c)
			_, e := uc.GetUsers(u2get)
			if e != nil {
				log.Println(e)
			}
		}
	default:
		fmt.Printf("not there yet")
	}

}


