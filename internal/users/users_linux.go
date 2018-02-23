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

package users

import (
	"bufio"
	"bytes"
	"fmt"
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

	// save the keys
	authkeys := u.AuthorizedKeys
	u = nu
	u.AuthorizedKeys = authkeys
	err = u.writeOutKeys()
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

	// start looking for processes

	return nil
}
