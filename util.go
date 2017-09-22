package main

import (
	//"github.com/ctdk/spqr/users"
	"github.com/ctdk/spqr/groups"
	"fmt"
	"log"
)

func convertUsersInterfaceSlice(u []interface{}) ([]*groups.Member, error) {
	log.Printf("u: %+v", u)
	l := len(u)
	users := make([]*groups.Member, l)
	for i, v := range u {
		s, ok := v.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("%v was supposed to be a map[string]interface{}, but was actually %T", v, v)
			return nil, err
		}
		m := new(groups.Member)
		m.Username = s["username"].(string)
		m.Status = s["status"].(string)
		users[i] = m
	}
	return users, nil
}
