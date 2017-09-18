package main

import (
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

func handleIncoming(keys []interface{}) {
	for _, k := range keys {
		switch k := k.(type) {
		case map[string]interface{}:
			fmt.Printf("what I expected: %+v\n", k)
			var payload string
			var handlingType uint8
			
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
		default:
			fmt.Printf("NOT what I expected: %T %v", k, k)
		}
	}
}
