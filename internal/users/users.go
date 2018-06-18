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
	"github.com/ctdk/spqr/internal/util"
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
	reenable bool
}

// New creates a new user. Some OS-specific constants are defined in the
// relevant go files for that platform.
func New(userName string, fullName string, homeDir string, shell string, action UserAction, groups []string, authorizedKeys []string) (*User, error) {
	// check for an existing user
	xu, _ := user.Lookup(userName)
	if xu != nil {
		err := fmt.Errorf("user %s already exists", userName)
		return nil, err
	}

	if homeDir == "" {
		homeDir = path.Join(DefaultHomeBase, userName)
	}
	if shell == "" {
		shell = DefaultShell
	}

	n := new(user.User)
	newUser := &User{n, nil, shell, action, groups, "", true, true, nil}
	newUser.Username = userName
	newUser.Name = fullName
	newUser.HomeDir = homeDir
	newUser.AuthorizedKeys = authorizedKeys

	return newUser, nil
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

	err = u.fillInGroups()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) Update() error {
	if !u.changed {
		return nil
	}

	return u.update()
}

func (u *User) Disable() error {
	err := u.deactivate()
	if err != nil {
		return err
	}

	err = u.deleteAuthKeys()
	if err != nil {
		return err
	}

	// OS-specific bits.
	err = u.disable()
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
		// Check for OS groups and create them if needed on systems
		// where that'll work.
		if err := u.reviewGroups(existingGroups); err != nil {
			return err
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
		
		err = u.setSshDirOwnership(authorizedKeyDir)
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

// get the user's shell and ssh keys
func (u *User) fillInUser() error {
	shell, err := getShell(u.Username)
	if err != nil {
		return err
	}
	u.Shell = shell

	authorizedKeyFile := u.authorizedKeyPath()

	authorizedKeys, err := getAuthorizedKeys(authorizedKeyFile)
	if err != nil {
		return err
	}

	u.AuthorizedKeys = authorizedKeys

	return nil
}

func (u *User) updateInfo(uEntry *UserInfo) error {
	r := u.Action == Disable && uEntry.Action == Create
	// Set the action, eh
	u.Action = uEntry.Action

	// bug out if the user's disabled or will be shortly
	if u.Action == Disable {
		return nil
	}

	uUp := new(userUpdated)

	// Only meaningful under Windows
	if r {
		uUp.reenable = r
		u.changed = true
	}

	// Low hanging fruit first - check if the shell and ssh keys need to be
	// changed.
	if uEntry.Shell != "" && u.Shell != uEntry.Shell {
		logger.Debugf("shell different for %s: o %s n %s", u.Username, u.Shell, uEntry.Shell)
		uUp.shell = uEntry.Shell
		u.changed = true
	}

	oldKeys, err := getAuthorizedKeys(u.authorizedKeyPath())
	if err != nil {
		return err
	}

	if !util.SliceEqual(oldKeys, uEntry.AuthorizedKeys) {
		logger.Debugf("authorized keys for %s didn't match", u.Username)
		uUp.authorizedKeys = uEntry.AuthorizedKeys
		u.changed = true
	}

	// The group updates may return here, but until that's sorted out
	// Windows-side, it's in the unix updateGroupInfo() method.

	err = u.updateGroupInfo(uEntry, uUp)

	if (uEntry.Name != "") && (uEntry.Name != u.Name) {
		logger.Debugf("Changing %s's full name from '%s' to '%s'.", u.Username, u.Name, uEntry.Name)
		uUp.name = uEntry.Name
		u.changed = true
	}

	if u.changed == true {
		logger.Debugf("user %s has information to update", u.Username)
		u.updated = uUp
	}

	return nil
}
