package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ctdk/spqr/groups"
	consul "github.com/hashicorp/consul/api"
	"log"
	"os/user"
	"strings"
)

var UserNotFound error = errors.New("That user was not found")

// temporary base here
const userKeyPrefix = "org/default/users"

type UserInfo struct {
	Username string `json:"username"`
	Name string `json:"name"`
	Groups []string `json:"groups"`
	HomeDir string `json:"home_dir"`
	Shell string `json:"shell"`
	Action UserAction `json:"action"`
	DoesNotExist bool `json:"does_not_exist"`
}

type UserConsulClient struct {
	*consul.Client
	userList []*groups.Member
	info []*UserInfo
}

func NewUserConsulClient(c *consul.Client) *UserConsulClient {
	return &UserConsulClient{c, []*groups.Member{}, []*UserInfo{},}
}

// get user information out of consul, get any that are present on the
func (client *UserConsulClient) GetUsers(userList []*groups.Member) ([]*User, error) {
	log.Println("in GetUsers")
	uinfo := make([]*UserInfo, 0, len(userList))
	client.userList = userList
	client.info = uinfo
	err := client.fetchInfo()
	if err != nil {
		return nil, err
	}
	for _, u := range client.info {
		fmt.Printf("a user info: %+v\n", u)
	}
	usarz := make([]*User, 0, len(client.info))
	for _, uEntry := range client.info {
		if uEntry.DoesNotExist {
			// A user needs to be created.
			n := new(user.User)
			newUser := &User{n, nil, uEntry.Shell, uEntry.Action, uEntry.Groups, true, true}
			newUser.Username = uEntry.Username
			newUser.Name = uEntry.Name
			newUser.HomeDir = uEntry.HomeDir
			newUser, err := New(uEntry.Username, uEntry.Name, uEntry.HomeDir, uEntry.Shell, uEntry.Action, uEntry.Groups)
			if err != nil {
				return nil, err
			}
			usarz = append(usarz, newUser)
		} else {
			// user already exists
			uObj, err := Get(uEntry.Username)
			if err != nil {
				return nil, err
			}
			usarz = append(usarz, uObj)
		}
	}

	return usarz, err
}

func (ui *UserInfo) populateUser() (*User, error) {
	u, err := Get(ui.Username)
	if err != nil { 
		// The user wasn't found
		return nil, err
	} else {
		// check for fields that have changed
		if u.Name != ui.Name {
			u.Name = ui.Name
			u.changed = true
		}
		if ui.HomeDir != "" && (u.HomeDir != ui.HomeDir) {
			u.HomeDir = ui.HomeDir
			u.changed = true
		}
	}
	return u, nil
}

func (ui *UserInfo) compareGroups(curGroups []string) bool {
	return false
}

func (c *UserConsulClient) UpdateUsers(userGaggle []*User) error {
	return nil
}

func (c *UserConsulClient) fetchInfo() error {
	kv := c.KV()
	for _, member := range c.userList {
		name := member.Username
		kval, _, err := kv.Get(strings.Join([]string{userKeyPrefix, name}, "/"), nil)
		if err != nil {
			return err
		}
		uInfo := new(UserInfo)
		err = json.Unmarshal(kval.Value, &uInfo)
		if err != nil {
			log.Printf("raw data was: '%s'", string(kval.Value))
			return err
		}
		if uInfo.Username == "" {
			uInfo.Username = uInfo.Name
		}
		log.Printf("uInfo status: %s group member status: %s", uInfo.Action, member.Status)
		if member.Status == groups.Disabled {
			uInfo.Action = Disable
		}
		uInfo.DoesNotExist = !userExists(uInfo.Username)
		fmt.Printf("got a user info: %+v\n", uInfo)
		c.info = append(c.info, uInfo)
	}

	return nil
}
