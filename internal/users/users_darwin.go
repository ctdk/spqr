// +build darwin

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

package users

import (
	"errors"
)

// no-op (+ error) at the moment

func (u *User) osCreateUser() error {
	return errors.New("user creation not supported on darwin")
}

func getShell(username string) (string, error) {
	return "", errors.New("getting a user's shell is not supported on darwin")
}

func osMakeNewGroup(groupName string) error {
	return errors.New("creating new groups is not supported on darwin")
}

func (u *User) killProcesses() error {
	return errors.New("killProcesses not implemented on darwin")
}

func (u *User) clearExtraGroups() error {
	return errors.New("clearExtraGroups not implemented on darwin")
}
func (u *User) updateGroups() error {
	return errors.New("updateGroups not implemented on darwin")
}
