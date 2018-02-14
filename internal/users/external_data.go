/*
 * Copyright (c) 2018, Jeremy Bingham (<jeremy@goiardi.gl>)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package users

import (
	"encoding/json"
	"github.com/ctdk/spqr/internal/groups"
	consul "github.com/hashicorp/consul/api"
	"github.com/tideland/golib/logger"
	"sort"
	"strings"
)

type UserExtDataClient struct {
	*consul.Client
	userList      []*groups.Member
	info          []*UserInfo
	userKeyPrefix string
}

func NewUserExtDataClient(c *consul.Client, userKeyPrefix string) *UserExtDataClient {
	return &UserExtDataClient{c, []*groups.Member{}, []*UserInfo{}, userKeyPrefix}
}

// get user information out of consul, get any that are present on the
func (c *UserExtDataClient) GetUsers(userList []*groups.Member) ([]*User, error) {
	logger.Debugf("In GetUsers")
	uinfo := make([]*UserInfo, 0, len(userList))
	c.userList = userList
	c.info = uinfo
	err := c.fetchInfo()
	if err != nil {
		return nil, err
	}
	usarz := make([]*User, 0, len(c.info))
	for _, uEntry := range c.info {
		logger.Debugf("Getting entry for %s. Does not exist? %v", uEntry.Username, uEntry.DoesNotExist)
		if uEntry.DoesNotExist {
			// A user needs to be created.
			newUser, err := New(uEntry.Username, uEntry.Name, uEntry.HomeDir, uEntry.Shell, uEntry.Action, uEntry.Groups, uEntry.AuthorizedKeys)
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
			err = uObj.updateInfo(uEntry)
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
	kv := c.KV()

	for _, member := range c.userList {
		name := member.Username
		kval, _, err := kv.Get(strings.Join([]string{c.userKeyPrefix, name}, "/"), nil)
		if err != nil {
			return err
		}

		if kval == nil {
			logger.Errorf("User '%s' not found under '%s'", name, c.userKeyPrefix)
			continue
		}

		uInfo := new(UserInfo)
		err = json.Unmarshal(kval.Value, &uInfo)
		if err != nil {
			return err
		}
		if uInfo.Username == "" {
			uInfo.Username = uInfo.Name
		}
		if member.Status == groups.Disabled {
			uInfo.Action = Disable
		}
		sort.Strings(uInfo.AuthorizedKeys)
		uInfo.DoesNotExist = !userExists(uInfo.Username)

		logger.Debugf("got a user info: %+v", uInfo)
		c.info = append(c.info, uInfo)
	}

	return nil
}
