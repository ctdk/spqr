// build +windows

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

package processes

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
)

var notImpErr = errors.New("Windows functionality is not implemented yet.")

func findUserProcesses(uid string) ([]*os.Process, error) {
	return nil, notImpErr
}

func killUserProcesses(username string) error {
	taskkillCmdPath, err := exec.LookPath("TASKKILL")
	if err != nil {
		return err
	}

	var taskkillArgs []string

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	taskkill := exec.Command(taskkillCmdPath, taskkillArgs...)
	taskkill.Stdout = &stdout
	taskkill.Stderr = &stderr

	if err := taskkill.Run(); err != nil {
		ferr := errors.New(strings.Join([]string{err.Error(), stderr.String()}, " -- "))
		return ferr
	}
	return nil
}
