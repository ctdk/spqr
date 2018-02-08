package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ctdk/spqr/groups"
	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	"log"
	"os/user"
	"strings"
)

var UserNotFound error = errors.New("That user was not found")

// temporary base here
const userKeyPrefix = "org/default/users"
const userVaultPrefix = "secret/spqr/users"

type UserInfo struct {
	Username string `json:"username"`
	Name string `json:"full_name"`
	Groups []string `json:"groups"`
	HomeDir string `json:"home_dir"`
	Shell string `json:"shell"`
	Action UserAction `json:"action"`
	DoesNotExist bool `json:"does_not_exist"`
	AuthorizedKeys []string
}

type UserExtDataClient struct {
	consul *consul.Client
	vault *vault.Client
	userList []*groups.Member
	info []*UserInfo
}

func NewUserExtDataClient(c *consul.Client, v *vault.Client) *UserExtDataClient {
	return &UserExtDataClient{c, v, []*groups.Member{}, []*UserInfo{},}
}

// get user information out of consul, get any that are present on the
func (client *UserExtDataClient) GetUsers(userList []*groups.Member) ([]*User, error) {
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

func (c *UserExtDataClient) UpdateUsers(userGaggle []*User) error {
	return nil
}

func (c *UserExtDataClient) fetchInfo() error {
	kv := c.consul.KV()
	vb := c.vault.Logical()

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
		
		// Look for authorized ssh public keys in vault
		if uInfo.Action != Disable && !uInfo.DoesNotExist {
			secret, err := vb.Read(strings.Join([]string{userVaultPrefix, name}, "/"))
			if err != nil {
				log.Printf("got an error: %s", err.Error())
			}
			log.Printf("hey, a secret: %+v", secret)
		}

		fmt.Printf("got a user info: %+v\n", uInfo)
		c.info = append(c.info, uInfo)
	}

	return nil
}

/*
	s, err := c.Logical().Read(strings.Join([]string{userVaultPrefix, "foo"}, "/"))
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
	}
	fmt.Printf("secret? %+v\n", s)
	data := s.Data
	fmt.Printf("hm: %T %v\n", data["ssh_keys"], data["ssh_keys"])
	j := make(map[string]interface{})
	e := json.Unmarshal([]byte(data["ssh_keys"].(string)), &j)
	if e != nil {
		fmt.Printf("e: %s\n", e.Error())
	}
	fmt.Printf("j: %T %+v\n", j, j)

###
secret? &{RequestID:9003c419-c5e1-4192-ad93-4e932c425c65 LeaseID: LeaseDuration:2764800 Renewable:false Data:map[ssh_keys:{"authorized_keys": ["ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAy/IqagcHKHWwk6iJNM/iVLAs+FWGYcx3KtJ1xyz8GgbvNf0NXXraDaAJewzxQAg+8V2E0/+6ynzzoyxSakaEeEKKI6PolHuGpKM44bG//8XZTesOiWE7W8KrpwhRSVkDy8zFsrmtIjinKjr0rYHX2Bw5FoXKjYWvbzXCsJhLOpbGOWkDNbLY5gL2nLvx5h5MO14ZKqHEm0eiQnB/b697Vqc4WvLZBCOra+0NKWcrJMHQGi5pijb9l1PlunmUche0Eo2l3J4F+TRzTMEfMZIsHM7Oa8LhHu+rwq6bdplTTykXUEUqHcBlE9IkY4uWZv7VRkaguuwwdlOlYW0/YM3ipQ== jeremy@nineveh.local"]}] Warnings:[] Auth:<nil> WrapInfo:<nil>}
hm: string {"authorized_keys": ["ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAy/IqagcHKHWwk6iJNM/iVLAs+FWGYcx3KtJ1xyz8GgbvNf0NXXraDaAJewzxQAg+8V2E0/+6ynzzoyxSakaEeEKKI6PolHuGpKM44bG//8XZTesOiWE7W8KrpwhRSVkDy8zFsrmtIjinKjr0rYHX2Bw5FoXKjYWvbzXCsJhLOpbGOWkDNbLY5gL2nLvx5h5MO14ZKqHEm0eiQnB/b697Vqc4WvLZBCOra+0NKWcrJMHQGi5pijb9l1PlunmUche0Eo2l3J4F+TRzTMEfMZIsHM7Oa8LhHu+rwq6bdplTTykXUEUqHcBlE9IkY4uWZv7VRkaguuwwdlOlYW0/YM3ipQ== jeremy@nineveh.local"]}
j: map[string]interface {} map[authorized_keys:[ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAy/IqagcHKHWwk6iJNM/iVLAs+FWGYcx3KtJ1xyz8GgbvNf0NXXraDaAJewzxQAg+8V2E0/+6ynzzoyxSakaEeEKKI6PolHuGpKM44bG//8XZTesOiWE7W8KrpwhRSVkDy8zFsrmtIjinKjr0rYHX2Bw5FoXKjYWvbzXCsJhLOpbGOWkDNbLY5gL2nLvx5h5MO14ZKqHEm0eiQnB/b697Vqc4WvLZBCOra+0NKWcrJMHQGi5pijb9l1PlunmUche0Eo2l3J4F+TRzTMEfMZIsHM7Oa8LhHu+rwq6bdplTTykXUEUqHcBlE9IkY4uWZv7VRkaguuwwdlOlYW0/YM3ipQ== jeremy@nineveh.local]]
###
*/
