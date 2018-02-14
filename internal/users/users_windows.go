// +build windows

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
