// +build windows

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

// windows specific functions and methods for creating users

// just stubs right now though

package users

import (
	"errors"
	"github.com/ctdk/spqr/internal/processes"
	"math/big"
	"os"
)

// Obviously these need to change drastically
const sshDirPerm = 0700
const authKeyPerm = 0644
const maxTmpDirNumBase int64 = 0xFFFFFFFF

var maxTmpDirNum *big.Int

const DefaultShell = "/bin/bash"
const DefaultHomeBase = "/home"

func init() {
	maxTmpDirNum = big.NewInt(maxTmpDirNumBase)
}

var notImpErr = errors.New("Windows functionality is not implemented yet.")

func osNew(userName string, fullName string, homeDir string, shell string, action UserAction, groups []string, authorizedKeys []string) (*User, error) {
	return nil, notImpErr
}

func (u *User) osCreateUser() error {
	return notImpErr
}

func (u *User) fillInUser() error {
	return notImpErr
}

func (u *User) update() error {
	return notImpErr
}

// rename?
func (u *User) setNoLogin() error {
	return notImpErr
}

func (u *User) changeShell(shell string) error {
	return notImpErr
}

func (u *User) updateInfo(uEntry *UserInfo) error {
	return notImpErr
}

func (u *User) getPrimaryGroup() (string, error) {
	return "", notImpErr
}

func getDefaultShell() string {
	return ""
}

func osMakeNewGroup(groupName string) error {
	return notImpErr
}

func getShell(username string) (string, error) {
	return "", notImpErr
}

func (u *User) clearExtraGroups() error {
	return notImpErr
}

func (u *User) updateGroups() error {
	return notImpErr
}

func (u *User) updateName() error {
	return notImpErr
}

func (u *User) createTempAuthKeyFile(baseDir string) (*os.File, error) {
	return nil, notImpErr
}

func (u *User) getUidGid() (int, int, error) {
	return 0, 0, notImpErr
}

func (u *User) killProcesses() error {
	return processes.KillUserProcesses(u.Uid)
}
