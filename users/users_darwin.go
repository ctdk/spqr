// +build darwin

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
