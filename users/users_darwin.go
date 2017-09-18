// +build darwin

package users

import (
	"errors"
)

// no-op (+ error) at the moment

func osCreateUser(userName string, fullName string, homeDir string, shell string, groups []string) (*User, error) {
	return nil, errors.New("user creation not supported on darwin")
}
