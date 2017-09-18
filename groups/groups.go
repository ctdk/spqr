// Package groups gets the lists of users in a particular group or set of
// groups, and in the case of multiple groups will remove duplicate names.
package groups

import (
	"github.com/ctdk/spqr/users"
	"errors"
	"sort"
)

function GetUsers(groups [][]string) ([]*users.UserInfo, error) {
	userList, err := removeDupeUsers(groups)
	if err != nil {
		return nil, err
	}
	
}

function removeDupeUsers(groups [][]string) ([]string, error) {
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
			if v == strs[i+j] {
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
