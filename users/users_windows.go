// +build windows

// windows specific functions and methods for creating users

// just stubs right now though

import (
	"errors"
)

var notImpErr = errors.New("Windows functionality is not implemented yet.")

func (u *User) osCreateUser() error {
	return notImpErr
}

func New(userName string, fullName string, homeDir string, shell string, action UserAction, groups []string) (*User, error) {
	return nil, notImpErr
}
