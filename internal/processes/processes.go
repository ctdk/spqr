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
	"github.com/tideland/golib/logger"
)

func KillUserProcesses(uid string) error {
	// Go through the process list twice, to be on the safe-ish side
	for i := 0; i < 2; i++ {
		// awfully OS specific, so...
		procs, err := findUserProcesses(uid)
		if err != nil {
			return err
		}
		logger.Debugf("Found %d processes for uid '%s' in round %d", len(procs), uid, i+1)

		if len(procs) == 0 {
			return nil
		}

		// now kill the processes
		for _, p := range procs {
			p.Kill()
		}
	}
	return nil
}
