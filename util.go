package main

import (
	//"github.com/ctdk/spqr/users"
	"github.com/ctdk/spqr/groups"
	"fmt"
)

func convertUsersInterfaceSlice(u []interface{}) ([]*groups.Member, error) {
	l := len(u)
	users := make([]*groups.Member, l)
	for i, v := range u {
		s, ok := v.(map[string]string)
		if !ok {
			err := fmt.Errorf("%v was supposed to be a map[string]string, but was actually %T", v)
			return nil, err
		}
		m := new(groups.Member)
		m.Username = s["username"]
		m.Status = s["status"]
		users[i] = m
	}
	return users, nil
}
