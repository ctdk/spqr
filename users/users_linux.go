// +build linux

// linux specific functionality to create users

package users

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func osCreateUser(userName string, fullName string, homeDir string, shell string, groups []string) (*User, error) {
	useraddPath, err := exec.LookPath("useradd")
	if err != nil {
		return nil, err
	}

	useraddArgs := []string{"-m", "-U", "-s", shell}

	if fullName != "" {
		useraddArgs = append(useraddArgs, []string{"-c", fullName}...)
	}
	
	if len(groups) > 0 {
		useraddArgs = append(useraddArgs, []string{"-G", strings.Join(groups, ",")}...)
	}

	if homeDir != "" {
		useraddArgs = append(useraddArgs, "-d", homeDir)
	}

	useraddArgs = append(useraddArgs, userName)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	
	useradd := exec.Command(useraddPath, useraddArgs...)
   	useradd.Stdout = &stdout
	useradd.Stderr = &stderr

	err = useradd.Run()
	if err != nil {
		ferr := fmt.Errorf("Error received while trying to create user: %s -- error from useradd program: %s", err.Error(), stderr.String())	
		return nil, ferr
	}

	u, err := Get(userName)

	if err != nil {
		return nil, err
	}

	return u, nil
}
