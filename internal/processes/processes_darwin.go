// build +linux

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
	"errors"
	"os"
)

func killUserProcesses(uid int) ([]*os.Process, error) {
	return nil, errors.New("Can't kill processes in darwin either (not sure how you managed to get here, for that matter.")
}

func findUserProcesses(uid int) ([]*os.Process, error) {
	return nil, errors.New("Can't look for processes in darwin either (not sure how you managed to get here, for that matter.")
}
