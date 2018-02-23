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
	"bytes"
	"github.com/tideland/golib/logger"
	"io"
	"os"
	"path"
	"strconv"
)

const (
	dentryToRead = 10
	statusBytes = 512
)

var statusUid = []byte("Uid:")

func findUserProcesses(uid string) ([]*os.Process, error) {
	// start looking for real
	procdir, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer procdir.Close()

	buid := []byte(uid)
	
	pids := make([]int, 0, 50)
	for {
		dentries, err := procdir.Readdir(dentryToRead)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		for _, d := range dentries {
			// only want directories
			if !d.IsDir() {
				continue
			}

			// also only want pid dirs
			n := d.Name()
			if n[0] < '0' || n[0] > '9' {
				continue
			}

			// Try and read the info from /proc/PID/stat. If it
			// fails the process may have disappeared or ended, so
			// there's no need to blow up.
			statusPath := path.Join(n, "status")
			status, err := os.Open(statusPath)
			if err != nil {
				logger.Debugf("Process status %s open failed, moving on: %s", err.Error())
				continue
			}
			defer status.Close()

			statusContents := make([]byte, statusBytes)
			var extraContent []byte
			var xuid []byte
			
			LineRead:
			for {
				nb, ferr := status.Read(statusContents[:statusBytes])
				if ferr != nil {
					if ferr == io.EOF {
						break
					}
					return nil, ferr
				}
				// taking the easy way out
				if extraContent != nil {
					statusContents = append(extraContent, statusContents[:nb]...)
					nb += len(extraContent)
					extraContent = nil
				}
				lines := bytes.Split(statusContents[:nb], []byte("\n"))

				for _, l := range lines {
					if bytes.Equal(l[:4], statusUid) {
						for i := 4; i < len(l); i++ {
							if l[i] >= '0' && l[i] <= '9' {
								for j := i; j < len(l); j++ {
									if l[j] < '0' || l[j] > '9' {
										xuid = l[i:j]
										break LineRead
									}
								}
							}
						}
					}
				}
			}
			// all done reading from /proc/PID/status, phew
			if bytes.Equal(buid, xuid) {
				pid, _ := strconv.Atoi(n)
				pids = append(pids, pid)
			}
		}
	}
	procs := make([]*os.Process, len(pids))
	for i, p := range pids {
		pr, _ := os.FindProcess(p) // it will always find a proc, even
					   // if it doesn't actually exist.
		procs[i] = pr
	}
	return procs, nil
}
