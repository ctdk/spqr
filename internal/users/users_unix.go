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
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/ctdk/spqr/internal/util"
	"github.com/tideland/golib/logger"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
)

const sshDirPerm = 0700
const authKeyPerm = 0644

const DefaultShell = "/bin/bash"
const DefaultHomeBase = "/home"

func (u *User) update() error {
	if u.updated.authorizedKeys != nil {
		if err := u.writeOutKeys(u.updated.authorizedKeys); err != nil {
			return err
		}
	}

	if u.updated.shell != "" {
		if u.Shell == "/sbin/nologin" {
			if err := u.passwdManipulate(false); err != nil {
				return err
			}
		}

		if err := u.changeShell(u.updated.shell); err != nil {
			return err
		}
	}

	if u.updated.groups != nil || u.updated.primaryGroup != "" {
		if err := u.updateGroups(); err != nil {
			return err
		}
	}

	if u.updated.name != "" {
		if err := u.updateName(); err != nil {
			return err
		}
	}

	return nil
}

// Golang and Windows groups don't seem to mesh real well yet, so moving the
// group specific (and any other Unixy specific bits) here.
func (u *User) disable() error {
	err = u.clearExtraGroups()
	if err != nil {
		return err
	}

	err = u.killProcesses()
	if err != nil {
		return err
	}
}

// NOTE: Anything involving UID/GID numbers, chown, and chmod will probably
// need to be abstracted out and reimplemented with OS-specific versions, even
// if the rest of the relevant methods can be shared between Unix and Windows.
func (u *User) createTempAuthKeyFile(baseDir string) (*os.File, error) {
	uid, gid, err := u.getUidGid()
	if err != nil {
		return nil, err
	}

	tmpAuthKeyFile, err := ioutil.TempFile(baseDir, "authorized_keys")
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

func (u *User) setSshDirOwnership(dir string) error {
	sshDir, err := os.Open(dir)
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
	return nil
}

// chsh might not be appropriate for dwarwin at least
// rename this?
func (u *User) deactivate() error {
	if err := u.passwdManipulate(true); err != nil {
		return err
	}
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

func (u *User) updateGroupInfo(uEntry *UserInfo, uUp *userUpdated) error {
	if !util.SliceEqual(uEntry.Groups, u.Groups) {
		logger.Debugf("groups didn't match for %s: o '%s' n '%s'", u.Username, strings.Join(u.Groups, ","), strings.Join(uEntry.Groups, ","))
		uUp.groups = uEntry.Groups
		u.changed = true
	}

	if (uEntry.PrimaryGroup != "") && (u.PrimaryGroup != uEntry.PrimaryGroup) {
		logger.Debugf("primary group for %s didn't match: o %s n %s", u.Username, u.PrimaryGroup, uEntry.PrimaryGroup)
		uUp.primaryGroup = uEntry.PrimaryGroup
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

func (u *User) reviewGroups(existingGroups map[string]bool) error {
	for _, g := range u.Groups {
		if !existingGroups[g] {
			if err := checkOrCreateGroup(g); err != nil {
				return err
			}
			existingGroups[g] = true
		}
	}
	return nil
}

func (u *User) fillInGroups() error {
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

func getDefaultShell() string {
	return "/bin/bash"
}
