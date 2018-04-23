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
	"bufio"
	"fmt"
	"github.com/tideland/golib/logger"
	"path"
	"os"
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
	PrimaryGroup   string
	changed        bool
	notExist       bool
	updated        *userUpdated
}

type UserInfo struct {
	Username       string     `json:"username"`
	Name           string     `json:"full_name"`
	Groups         []string   `json:"groups"`
	PrimaryGroup   string     `json:"primary_group"`
	HomeDir        string     `json:"home_dir"`
	Shell          string     `json:"shell"`
	Action         UserAction `json:"action"`
	DoesNotExist   bool       `json:"does_not_exist"`
	AuthorizedKeys []string   `json:"authorized_keys"`
}

type userUpdated struct {
	name           string
	groups         []string
	primaryGroup   string
	shell          string
	authorizedKeys []string
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

	u := &User{osUser, nil, "", NullAction, nil, "", false, false, nil}

	err = u.fillInUser()
	if err != nil {
		return nil, err
	}

	pg, err := u.getPrimaryGroup()
	logger.Debugf("primary group for %s: '%s'", u.Username, pg)
	if err != nil {
		return nil, err
	}
	u.PrimaryGroup = pg

	gids, _ := u.GroupIds()

	for _, g := range gids {
		gr, _ := user.LookupGroupId(g)
		if gr.Name == u.PrimaryGroup {
			continue
		}
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

	err = u.clearExtraGroups()
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
				if err := checkOrCreateGroup(g); err != nil {
					return err
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
			if !u.notExist {
				err := u.Disable()
				if err != nil {
					return err
				}
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

func checkOrCreateGroup(name string) error {
	logger.Debugf("looking up group %s", name)
	gPresent, _ := user.LookupGroup(name)
	if gPresent == nil {
		err := MakeNewGroup(name)
		if err != nil {
			return err
		}
	}
	return nil
}

// Looks like these methods are actually suitable for Windows running an SSH
// server after all, since that's rapidly becoming a thing.

// NOTE: Anything involving UID/GID numbers, chown, and chmod will probably
// need to be abstracted out and reimplemented with OS-specific versions, even
// if the rest of the relevant methods can be shared between Unix and Windows.

func userExists(userName string) bool {
	u, _ := user.Lookup(userName)
	if u != nil {
		return true
	}
	return false
}

func getAuthorizedKeys(authorizedKeyFile string) ([]string, error) {
	var authorizedKeys []string

	if aKeys, err := os.Open(authorizedKeyFile); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		defer aKeys.Close()
		authKeys := bufio.NewScanner(aKeys)
		for authKeys.Scan() {
			authorizedKeys = append(authorizedKeys, authKeys.Text())
		}
		if err = authKeys.Err(); err != nil {
			return nil, err
		}
	}

	sort.Strings(authorizedKeys)

	return authorizedKeys, nil
}

func (u *User) writeOutKeys(authorizedKeys []string) error {
	logger.Debugf("writing out authorized keys for %s", u.Username)
	if len(authorizedKeys) == 0 {
		err := fmt.Errorf("no SSH keys given for %s", u.Username)
		return err
	}

	authorizedKeyFile := u.authorizedKeyPath()
	authorizedKeyDir := path.Dir(authorizedKeyFile)

	if _, err := os.Stat(authorizedKeyDir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		err = os.Mkdir(authorizedKeyDir, sshDirPerm)
		if err != nil {
			return err
		}
		sshDir, err := os.Open(authorizedKeyDir)
		if err != nil {
			return err
		}
		uid, gid, err := u.getUidGid()
		if err != nil {
			return err
		}

		err = sshDir.Chown(uid, gid)
		if err != nil {
			return err
		}
	}

	tmpAuthKeys, err := u.createTempAuthKeyFile(authorizedKeyDir)
	if err != nil {
		return err
	}

	for _, l := range authorizedKeys {
		_, err = tmpAuthKeys.WriteString(l)
		if err != nil {
			tmpAuthKeys.Close()
			return err
		}
		_, err = tmpAuthKeys.WriteString("\n")
		if err != nil {
			tmpAuthKeys.Close()
			return err
		}
	}

	tmpAuthKeyPath := tmpAuthKeys.Name()
	tmpAuthKeys.Close()
	err = os.Rename(tmpAuthKeyPath, authorizedKeyFile)
	if err != nil {
		return err
	}
	logger.Debugf("successfully wrote authorized keys for %s", u.Username)
	return nil
}

func (u *User) deleteAuthKeys() error {
	authKeyPath := u.authorizedKeyPath()
	err := os.Remove(authKeyPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	logger.Debugf("deleted authorized keys for %s", u.Username)
	return nil
}

func (u *User) authorizedKeyPath() string {
	return path.Join(u.HomeDir, ".ssh", "authorized_keys")
}
