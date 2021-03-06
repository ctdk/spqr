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

// Package groups gets the lists of users in a particular group or set of
// groups, and in the case of multiple groups will remove duplicate names.
package groups

import (
	"errors"
	"fmt"
	"github.com/ctdk/spqr/internal/util"
	"github.com/tideland/golib/logger"
	"sort"
)

type Member struct {
	Username string `json:"username"`
	Status   string `json:"status"`
	CommonGroups []string `json:"common_groups"`
}

const (
	Enabled  = "enabled"
	Disabled = "disabled"
)

type GroupMembers []*Member

func (gm GroupMembers) Len() int           { return len(gm) }
func (gm GroupMembers) Swap(i, j int)      { gm[i], gm[j] = gm[j], gm[i] }
func (gm GroupMembers) Less(i, j int) bool { return gm[i].Username < gm[j].Username }

func RemoveDupeUsers(groups [][]*Member) ([]*Member, error) {
	var list []*Member

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
		list = make([]*Member, 0, listCap)
		for _, l := range groups {
			list = append(list, l...)
		}
	}

	// Remove the dupes now. Even just one group can have duplicate entries,
	// so snip them out regardless of list length.
	sort.Sort(GroupMembers(list))

	var listSort string
	for w, q := range list {
		listSort = fmt.Sprintf("%s %d %+v", listSort, w, q)
	}
	logger.Debugf("sorted list: %v", listSort)

	// borrowing from goiardi some here
	for i, u := range list {
		logger.Debugf("user in RemoveDupeUsers: %+v", u)
		if i > len(list) {
			break
		}
		j := 1
		s := 0
		for {
			if i+j >= len(list) {
				break
			}
			if u.Username == list[i+j].Username {
				j++
				s++
				u.CommonGroups = append(u.CommonGroups, list[i+j].CommonGroups...)
			} else {
				break
			}
		}
		if s == 0 {
			continue
		}
		if u.Status != Enabled {
			for z := i + 1; z < (i+s)-1; z++ {
				if list[z].Status == Enabled {
					u.Status = Enabled
					break
				}
			}
		}

		list = delTwoPosElements(i+1, s, list)
	}

	for _, gu := range list {
		sort.Strings(gu.CommonGroups)
		gu.CommonGroups = util.RemoveDupeSliceString(gu.CommonGroups)
	}

	listSort = ""
	for w, q := range list {
		listSort = fmt.Sprintf("%s %d %+v", listSort, w, q)
	}
	logger.Debugf("the sorted and de-duped list: %v", listSort)
	return list, nil
}

// borrowing from some goiardi work here too
func delTwoPosElements(pos int, skip int, gm []*Member) []*Member {
	gm = append(gm[:pos], gm[pos+skip:]...)
	return gm
}
