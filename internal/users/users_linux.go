// +build linux

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

// linux specific functionality to create users
// NB: Most/everything in this file is probably applicable to most if not all
// non-darwin Unixes, so the file may need moved to a new name (with any truly
// linux-specific items moved back into users_linux.go accordingly).

package users

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ctdk/spqr/internal/processes"
	"github.com/tideland/golib/logger"
	"os"
	"os/exec"
	"strings"
)

func (u *User) osCreateUser() error {
	useraddPath, err := exec.LookPath("useradd")
	if err != nil {
		return err
	}

	useraddArgs := []string{"-m", "-U", "-s", u.Shell}

	if u.Name != "" {
		useraddArgs = append(useraddArgs, []string{"-c", u.Name}...)
	}

	if len(u.Groups) > 0 {
		useraddArgs = append(useraddArgs, []string{"-G", strings.Join(u.Groups, ",")}...)
	}

	if u.PrimaryGroup != "" {
		useraddArgs = append(useraddArgs, []string{"-g", u.PrimaryGroup}...)
	}

	if u.HomeDir != "" {
		useraddArgs = append(useraddArgs, "-d", u.HomeDir)
	}

	useraddArgs = append(useraddArgs, u.Username)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	useradd := exec.Command(useraddPath, useraddArgs...)
	useradd.Stdout = &stdout
	useradd.Stderr = &stderr

	err = useradd.Run()
	if err != nil {
		ferr := fmt.Errorf("Error received while trying to create user: %s -- error from useradd program: %s", err.Error(), stderr.String())
		return ferr
	}

	nu, err := Get(u.Username)

	if err != nil {
		return err
	}

	authKeys := u.AuthorizedKeys
	u = nu

	// save the keys
	err = u.writeOutKeys(authKeys)
	if err != nil {
		return err
	}

	return nil
}

func osMakeNewGroup(groupName string) error {
	groupaddPath, err := exec.LookPath("groupadd")
	if err != nil {
		return err
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	groupadd := exec.Command(groupaddPath, groupName)
	groupadd.Stdout = &stdout
	groupadd.Stderr = &stderr
	err = groupadd.Run()
	if err != nil {
		return fmt.Errorf("Error received trying to create group %s: %s %s", groupName, err.Error(), stderr.String())
	}
	return nil
}

func getShell(username string) (string, error) {
	var shell string
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		return "", err
	}
	defer passwd.Close()
	pl := bufio.NewScanner(passwd)
	for pl.Scan() {
		line := pl.Text()
		if strings.HasPrefix(line, fmt.Sprintf("%s:", username)) {
			fields := strings.Split(line, ":")
			shell = fields[len(fields)-1]
			break
		}
	}
	if err = pl.Err(); err != nil {
		return "", err
	}
	return shell, nil
}

func (u *User) killProcesses() error {
	if u.Uid == "0" {
		return fmt.Errorf("Will not kill processes for uid 0")
	}

	// kill the processes
	return processes.KillUserProcesses(u.Uid)
}

func (u *User) clearExtraGroups() error {
	// inside docker at least 'groupmems' required a password to add/remove
	// users from a group. Weeeeeird.
	// Bail early if the user is already not in any extra groups
	if len(u.Groups) == 0 {
		return nil
	}
	uUp := new(userUpdated)
	uUp.groups = make([]string, 0)
	u.updated = uUp
	logger.Debugf("Removing %s from all extra groups", u.Username)
	return u.updateGroups()
}

func (u *User) updateName() error {
	userModArgs := []string{"-c", u.updated.name}
	logger.Debugf("Updating full name for %s to '%s'", u.Username, u.updated.name)
	return u.runUserMod(userModArgs)
}

func (u *User) updateGroups() error {
	userModArgs := make([]string, 0, 4)

	if u.updated.groups != nil {
		for _, g := range u.updated.groups {
			if err := checkOrCreateGroup(g); err != nil {
				return err
			}
		}
		ua := []string{"-G", strings.Join(u.updated.groups, ",")}
		logger.Debugf("Updating groups for %s to '%s'", u.Username, strings.Join(u.updated.groups, ","))
		userModArgs = append(userModArgs, ua...)
	}

	if u.updated.primaryGroup != "" {
		ua := []string{"-g", u.updated.primaryGroup}
		logger.Debugf("Updating primary group for %s to '%s'", u.Username, u.updated.primaryGroup)
		userModArgs = append(userModArgs, ua...)
	}

	logger.Debugf("running usermod on '%s' with these arguments: %s", u.Username, strings.Join(userModArgs, " "))
	return u.runUserMod(userModArgs)
}

func (u *User) runUserMod(userModArgs []string) error {
	usermodPath, err := exec.LookPath("usermod")
	if err != nil {
		return err
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	userModArgs = append(userModArgs, u.Username)
	usermod := exec.Command(usermodPath, userModArgs...)
	usermod.Stdout = &stdout
	usermod.Stderr = &stderr

	err = usermod.Run()
	if err != nil {
		return fmt.Errorf("Error received while modifying %s: %s :: %s", u.Username, err.Error(), stderr.String())
	}
	return nil
}

func (u *User) passwdManipulate(lock bool) error {
	pPath, err := exec.LookPath("passwd") // can't imagine that would fail
	if err != nil {
		return err
	}

	var op string
	if lock {
		op = "-l"
	} else {
		op = "-u"
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	p := exec.Command(pPath, op, u.Username)
	p.Stdout = &stdout
	p.Stderr = &stderr

	err = p.Run()
	if err != nil {
		if !lock && !strings.Contains(stderr.String(), "passwordless account") {
			return fmt.Errorf("Error received while locking/unlocking account %s: %s :: %s || l %v c %v a %v", u.Username, err.Error(), stderr.String(), !lock, strings.Contains(stderr.String(), "passwordless account"), !(!lock && !strings.Contains(stderr.String(), "passwordless account")))
		}
	}

	return nil
}
