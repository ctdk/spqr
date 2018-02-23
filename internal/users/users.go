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

// Package users contains methods for creating and managing users and their
// ssh keys.
package users

import (
	"fmt"
	"github.com/tideland/golib/logger"
	"os/user"
	"sort"
)

type UserAction string

const (
	NullAction UserAction = "null"
	Create                = "create"
	Disable               = "disable"
)

type User struct {
	*user.User
	AuthorizedKeys []string
	Shell          string
	Action         UserAction
	Groups         []string
	changed        bool
	notExist       bool
}

type UserInfo struct {
	Username       string     `json:"username"`
	Name           string     `json:"full_name"`
	Groups         []string   `json:"groups"`
	HomeDir        string     `json:"home_dir"`
	Shell          string     `json:"shell"`
	Action         UserAction `json:"action"`
	DoesNotExist   bool       `json:"does_not_exist"`
	AuthorizedKeys []string   `json:"authorized_keys"`
}

// New creates a new user. It's a pass-through to an OS-specific function, see
// the appropriate one for details.
func New(userName string, fullName string, homeDir string, shell string, action UserAction, groups []string, authorizedKeys []string) (*User, error) {
	return osNew(userName, fullName, homeDir, shell, action, groups, authorizedKeys)
}

// Get a user, if it exists.
func Get(username string) (*User, error) {
	osUser, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}

	u := &User{osUser, nil, "", NullAction, nil, false, false}

	err = u.fillInUser()
	if err != nil {
		return nil, err
	}

	gids, _ := u.GroupIds()

	for _, g := range gids {
		gr, _ := user.LookupGroupId(g)
		u.Groups = append(u.Groups, gr.Name)
	}
	sort.Strings(u.Groups)
	
	return u, nil
}

func (u *User) Update() error {
	if !u.changed {
		return nil
	}

	return u.update()
}

func (u *User) Disable() error {
	err := u.setNoLogin()
	if err != nil {
		return err
	}

	err = u.deleteAuthKeys()
	if err != nil {
		return err
	}

	err = u.killProcesses()
	if err != nil {
		return err
	}

	return nil
}

func MakeNewGroup(groupName string) error {
	logger.Debugf("Making new group %s", groupName)
	return osMakeNewGroup(groupName)
}

// ProcessUsers updates or create users as needed.
func ProcessUsers(userList []*User) error {
	existingGroups := make(map[string]bool)

	for _, u := range userList {
		// Check for OS groups and create them if needed
		for _, g := range u.Groups {
			if !existingGroups[g] {
				logger.Debugf("looking up group %s", g)
				gPresent, _ := user.LookupGroup(g)
				if gPresent == nil {
					err := MakeNewGroup(g)
					if err != nil {
						return nil
					}
				}
				existingGroups[g] = true
			}
		}

		if u.notExist && u.Action != Disable {
			err := u.osCreateUser()
			if err != nil {
				uerr := fmt.Errorf("Error attempting to create user %s: %s", u.Username, err.Error())
				return uerr
			}
		} else if u.Action == Disable {
			err := u.Disable()
			if err != nil {
				return err
			}
		} else {
			err := u.Update()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
