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

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ctdk/spqr/config"
	"github.com/ctdk/spqr/internal/groups"
	"github.com/ctdk/spqr/internal/state"
	"github.com/ctdk/spqr/internal/users"
	consul "github.com/hashicorp/consul/api"
	"github.com/tideland/golib/logger"
)

const (
	notAThing uint8 = iota
	keyPrefix
	consulEvent
)

var handleDesc = map[uint8]string{
	keyPrefix:   "key prefix",
	consulEvent: "event",
}

func handleIncoming(c *consul.Client, stateHolder *state.State, incomingCh chan *state.Indices, keys []interface{}) {
	var handlingType uint8
	var groupLists [][]*groups.Member

	idxIncoming := make([]*state.Indices, 0, len(keys))
	logger.Debugf("Number of keys incoming: %d", len(keys))

	for _, k := range keys {
		switch k := k.(type) {
		case map[string]interface{}:
			logger.Debugf("what I expected: %+v", k)
			var payload string

			if stateHolder != nil {
				cidx, _ := k["CreateIndex"].(json.Number).Int64()
				midx, _ := k["ModifyIndex"].(json.Number).Int64()
				lidx, _ := k["LockIndex"].(json.Number).Int64()
				if !stateHolder.DoProcessIncoming(cidx, midx) {
					continue
				}
				idx := new(state.Indices)
				idx.CreateIndex = cidx
				idx.ModifyIndex = midx
				idx.LockIndex = lidx
				idxIncoming = append(idxIncoming, idx)
			} else {
				logger.Debugf("erm, stateHolder is nil, but shouldn't be.")
			}

			// within one request, everything will be just one kind
			// of thing so this only needs to be checked once.
			if handlingType == notAThing {
				if r, ok := k["Value"].(string); ok {
					payload = r
					handlingType = keyPrefix
				} else if r, ok := k["Payload"].(string); ok {
					payload = r
					handlingType = consulEvent
				} else {
					logger.Warningf("doesn't appear to be a keyprefix or event, moving on")
					continue
				}
			} else {
				var ok bool
				switch handlingType {
				case keyPrefix:
					payload, ok = k["Value"].(string)
				case consulEvent:
					payload, ok = k["Payload"].(string)
				default:
					logger.Fatalf("there's no way this should be reachable")
				}
				if !ok {
					logger.Errorf("expected a string, but something went wrong")
				}
			}

			val, err := base64.StdEncoding.DecodeString(payload)
			if err != nil {
				logger.Errorf(err.Error())
			}

			j := make(map[string]interface{})
			err = json.Unmarshal([]byte(val), &j)
			if err != nil {
				logger.Errorf(err.Error())
				continue
			}
			logger.Debugf("this is a %s", handleDesc[handlingType])
			switch handlingType {
			case keyPrefix:
				convUsers, err := convertUsersInterfaceSlice(j["members"].([]interface{}))
				if err != nil {
					logger.Errorf(err.Error())
					continue
				}
				groupLists = append(groupLists, convUsers)
			default:
				logger.Debugf("not handling %s yet in switch", handleDesc[handlingType])
			}
		default:
			logger.Errorf("NOT what I expected: %T %v", k, k)
		}
	}

	// So what do we do?
	switch handlingType {
	case keyPrefix:
		if u2get, err := groups.RemoveDupeUsers(groupLists); err != nil {
			logger.Errorf(err.Error())
		} else {
			uc := users.NewUserExtDataClient(c, config.Config.UserKeyPrefix)
			usarz, e := uc.GetUsers(u2get)
			if e != nil {
				logger.Errorf(e.Error())
			}
			perr := users.ProcessUsers(usarz)
			if perr != nil {
				logger.Errorf(perr.Error())
			}
		}
	default:
		logger.Infof("not handling events (or anything else besides key prefix watches) yet")
	}

	// Send any index updates to the state to process
	if stateHolder != nil {
		for _, idx := range idxIncoming {
			incomingCh <- idx
		}
		close(incomingCh)
	}
}

func convertUsersInterfaceSlice(u []interface{}) ([]*groups.Member, error) {
	l := len(u)
	users := make([]*groups.Member, l)
	for i, v := range u {
		s, ok := v.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("%v was supposed to be a map[string]interface{}, but was actually %T", v, v)
			return nil, err
		}
		m := new(groups.Member)
		m.Username = s["username"].(string)
		m.Status = s["status"].(string)
		users[i] = m
	}
	return users, nil
}
