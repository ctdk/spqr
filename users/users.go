// Package users contains methods for creating and managing users and their
// ssh keys.
package users

import (
	"log"
	"os/user"
)

const DefaultShell = "/bin/bash"
const DefaultHomeBase = "/home"

type UserAction uint8

const (
	NullAction UserAction = iota
	Create
	Disable
)

type User struct {
	*user.User
	SSHKeys []string
	Shell string
	Action UserAction
}

// Get a user, if it exists.
func Get(username string) (*User, error) {
	osUser, err := user.Lookup(username)
	log.Printf("osUser? %+v", osUser)
	if err != nil {
		return nil, err
	}
	u := &User{osUser,nil,""}
	err = u.fillInUser()

	if err != nil {
		return nil, err
	}
	return u, nil
}

func (u *User) Update() error {
	return u.update()
}

func (u *User) Disable() error {
	// At the moment, removing their ssh keys is probably sufficient. Down
	// the road, killing all their processes may be in order.
	return u.deleteAuthKeys()
}
