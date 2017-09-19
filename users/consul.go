package users

import (
	"encoding/json"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"log"
	"strings"
)

// temporary base here
const userKeyPrefix = "org/default/users"

type userInfo struct {
	Name string `json:"name"`
	FullName string `json:"full_name"`
	Groups []string `json:"groups"`
	HomeDir string `json:"home_dir"`
	Shell string `json:"shell"`
	Action UserAction `json:"action"`
}

type UserConsulClient struct {
	*consul.Client
	userList []string
	info []*userInfo
}

func NewUserConsulClient(c *consul.Client) *UserConsulClient {
	return &UserConsulClient{c, []string{}, []*userInfo{},}
}

// get user information out of consul, get any that are present on the
func (client *UserConsulClient) GetUsers(userList []string) ([]*User, error) {
	log.Println("in GetUsers")
	uinfo := make([]*userInfo, 0, len(userList))
	client.userList = userList
	client.info = uinfo
	err := client.fetchInfo()
	if err != nil {
		return nil, err
	}
	for _, u := range client.info {
		fmt.Printf("a user info: %+v\n", u)
	}

	return nil, err
}

func (c *UserConsulClient) UpdateUsers(userGaggle []*User) error {
	return nil
}

func (c *UserConsulClient) fetchInfo() error {
	kv := c.KV()
	for _, name := range c.userList {
		kval, _, err := kv.Get(strings.Join([]string{userKeyPrefix, name}, "/"), nil)
		if err != nil {
			return err
		}
		uInfo := new(userInfo)
		err = json.Unmarshal(kval.Value, &uInfo)
		if err != nil {
			log.Printf("raw data was: '%s'", string(kval.Value))
			return err
		}
		fmt.Printf("got a user info: %+v\n", uInfo)
		c.info = append(c.info, uInfo)
	}

	return nil
}

