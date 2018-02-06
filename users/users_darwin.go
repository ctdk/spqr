// +build darwin

package users

import (
	"errors"
)

// no-op (+ error) at the moment

func (u *User) osCreateUser() error {
	return errors.New("user creation not supported on darwin")
}
