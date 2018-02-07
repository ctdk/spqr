// +build linux

// linux specific functionality to create users

package users

import (
	"bytes"
	"fmt"
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
	u = nu

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
