package main

import (
	//"github.com/ctdk/spqr/users"
	"errors"
	"fmt"
	"sort"
)

/*
func getUsers(groups [][]string) ([]*users.UserInfo, error) {
	userList, err := removeDupeUsers(groups)
	_ = userList
	if err != nil {
		return nil, err
	}
	return nil, err
}
*/

func removeDupeUsers(groups [][]string) ([]string, error) {
	var list []string

	if len(groups) == 1 { // just one group
		list = groups[0]
	} else if len(groups) == 0 {
		err := errors.New("no groups of users actually provided")
		return nil, err
	} else {
		listCap := 0
		for _, y := range groups {
			listCap += len(y)
		}
		list = make([]string, 0, listCap)
		for _, l := range groups {
			list = append(list, l...)
		}
	}

	// Remove the dupes now. Even just one group can have duplicate entries,
	// so snip them out regardless of list length.
	sort.Strings(list)

	// borrowing from goiardi some here
	for i, u := range list {
		if i > len(list) {
			break
		}
		j := 1
		s := 0
		for {
			if i+j >= len(list) {
				break
			}
			if u == list[i+j] {
				j++
				s++
			} else {
				break
			}
		}
		if s == 0 {
			continue
		}
		list = delTwoPosElements(i+1, s, list)
	}
	return list, nil
}

// borrowing from some goiardi work here too
func delTwoPosElements(pos int, skip int, strs []string) []string {
	strs = append(strs[:pos], strs[pos+skip:]...)
	return strs
}

func convertUsersInterfaceSlice(u []interface{}) ([]string, error) {
	l := len(u)
	users := make([]string, l)
	for i, v := range u {
		s, ok := v.(string)
		if !ok {
			err := fmt.Errorf("%v was supposed to be a string, but was actually %T", v)
			return nil, err
		}
		users[i] = s
	}
	return users, nil
}
