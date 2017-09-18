package users

import (

)

type userInfo struct {
	Name string
	FullName string
	Groups []string
	HomeDir string
	Shell string
	Action UserAction
}

type userConsulClient struct {
	*consul.Client
	userList []string
	info []*userInfo
}

// get user information out of consul, get any that are present on the
func GetUsers(c *consul.Client, userList []string) ([]*User, error) {
	client := &userConsulClient{c, userList, []*userInfo{},}
	err := client.fetchInfo()
	if err != nil {
		return nil, err
	}


}

func (c *userConsulClient) fetchInfo() error {
	kv := c.KV()
	
}
