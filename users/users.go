// Package users contains methods for creating and managing users and their
// ssh keys.
package users

import (
	"fmt"
	"log"
	"os/user"
)

type UserAction string

const (
	NullAction UserAction = "null"
	Create = "create"
	Disable = "disable"
)

type User struct {
	*user.User
	SSHKeys []string
	Shell string
	Action UserAction
	Groups []string
	changed bool
	notExist bool
}

// Get a user, if it exists.
func Get(username string) (*User, error) {
	osUser, err := user.Lookup(username)
	log.Printf("osUser? %+v", osUser)
	if err != nil {
		return nil, err
	}
	u := &User{osUser, nil, "", NullAction, nil, false, false}
	err = u.fillInUser()
	u.Groups, _ = u.GroupIds()

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

// ProcessUsers updates or create users as needed.
func ProcessUsers(userList []*User) error {
	for _, u := range userList {
		if u.notExist {
			err := u.osCreateUser()
			if err != nil {
				uerr := fmt.Errorf("Error attempting to create user %s: %s", u.Username, err.Error())
				return uerr
			}
		} else {

		}
	}
	return nil
}
