// +build linux darwin

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

// common functions shared across whatever unixes this might someday support

package users

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/ctdk/spqr/internal/util"
	"github.com/tideland/golib/logger"
	"math/big"
	"os"
	"os/exec"
	"os/user"
	"path"
	"sort"
	"strconv"
	"strings"
)

const sshDirPerm = 0700
const authKeyPerm = 0644
const maxTmpDirNumBase int64 = 0xFFFFFFFF

var maxTmpDirNum *big.Int

const DefaultShell = "/bin/bash"
const DefaultHomeBase = "/home"

func init() {
	maxTmpDirNum = big.NewInt(maxTmpDirNumBase)
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

func (u *User) update() error {
	// TODO: update other fields besides just SSH keys, like shell, various
	// /etc/passwd entries, etc.
	if err := u.writeOutKeys(); err != nil {
		return err
	}

	if err := u.changeShell(u.Shell); err != nil {
		return err
	}

	if err := u.updateGroups(); err != nil {
		return err
	}

	return nil
}

func (u *User) writeOutKeys() error {
	logger.Debugf("writing out authorized keys for %s", u.Username)
	if len(u.AuthorizedKeys) == 0 {
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

	for _, l := range u.AuthorizedKeys {
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

func osNew(userName string, fullName string, homeDir string, shell string, action UserAction, groups []string, authorizedKeys []string) (*User, error) {
	if homeDir == "" {
		homeDir = path.Join(DefaultHomeBase, userName)
	}
	if shell == "" {
		shell = DefaultShell
	}

	// check for an existing user
	xu, _ := user.Lookup(userName)
	if xu != nil {
		err := fmt.Errorf("user %s already exists", userName)
		return nil, err
	}

	n := new(user.User)
	newUser := &User{n, nil, shell, action, groups, "", true, true}
	newUser.Username = userName
	newUser.Name = fullName
	newUser.HomeDir = homeDir
	newUser.AuthorizedKeys = authorizedKeys

	return newUser, nil
}

func (u *User) createTempAuthKeyFile(baseDir string) (*os.File, error) {
	uid, gid, err := u.getUidGid()
	if err != nil {
		return nil, err
	}

	n, err := rand.Int(rand.Reader, maxTmpDirNum)
	if err != nil {
		return nil, err
	}

	tmpAuthKeyPath := path.Join(baseDir, strings.Join([]string{"authorized_keys", n.String()}, "-"))
	tmpAuthKeyFile, err := os.Create(tmpAuthKeyPath)
	if err != nil {
		return nil, err
	}
	err = tmpAuthKeyFile.Chmod(authKeyPerm)
	if err != nil {
		return nil, err
	}
	err = tmpAuthKeyFile.Chown(uid, gid)
	if err != nil {
		return nil, err
	}

	return tmpAuthKeyFile, nil
}

func (u *User) getUidGid() (int, int, error) {
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, 0, err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return 0, 0, err
	}
	return uid, gid, nil
}

func userExists(userName string) bool {
	u, _ := user.Lookup(userName)
	if u != nil {
		return true
	}
	return false
}

// chsh might not be appropriate for dwarwin at least
func (u *User) setNoLogin() error {
	return u.changeShell("/sbin/nologin")
}

func (u *User) changeShell(shell string) error {
	chshPath, err := exec.LookPath("chsh")
	if err != nil {
		return err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// make sure RHEL/CentOS or some other sort of unix doesn't use
	// something else besides /sbin/nologin for setting an account to
	// be unable to login.
	chsh := exec.Command(chshPath, "-s", shell, u.Username)
	chsh.Stdout = &stdout
	chsh.Stderr = &stderr
	err = chsh.Run()
	if err != nil {
		return fmt.Errorf("Error received trying to user %s to %s: %s %s", u.Username, shell, err.Error(), stderr.String())
	}

	return nil
}

func (u *User) updateInfo(uEntry *UserInfo) error {
	// Set the action, eh
	u.Action = uEntry.Action

	// bug out if the user's disabled or will be shortly
	if u.Action == Disable {
		return nil
	}

	// Low hanging fruit first - check if the shell and ssh keys need to be
	// changed.
	if uEntry.Shell != "" && u.Shell != uEntry.Shell {
		logger.Debugf("shell different for %s: o %s n %s", u.Username, u.Shell, uEntry.Shell)
		u.Shell = uEntry.Shell
		u.changed = true
	}

	oldKeys, err := getAuthorizedKeys(u.authorizedKeyPath())
	if err != nil {
		return err
	}
	if !util.SliceEqual(oldKeys, uEntry.AuthorizedKeys) {
		logger.Debugf("authorized keys for %s didn't match", u.Username)
		u.AuthorizedKeys = uEntry.AuthorizedKeys
		u.changed = true
	}

	if !util.SliceEqual(uEntry.Groups, u.Groups) {
		logger.Debugf("groups didn't match for %s: o '%s' n '%s'", u.Username, strings.Join(u.Groups, ","), strings.Join(uEntry.Groups, ","))
		u.Groups = uEntry.Groups
		u.changed = true
	}

	if (uEntry.PrimaryGroup != "") && (u.PrimaryGroup != uEntry.PrimaryGroup) {
		logger.Debugf("primary group for %s didn't match: o %s n %s", u.Username, u.PrimaryGroup, uEntry.PrimaryGroup)
		u.PrimaryGroup = uEntry.PrimaryGroup
		u.changed = true
	}


	return nil
}

func (u *User) getPrimaryGroup() (string, error) {
	gr, err := user.LookupGroupId(u.Gid)
	if err != nil {
		return "", err
	}
	return gr.Name, nil
}

func getDefaultShell() string {
	return "/bin/bash"
}
