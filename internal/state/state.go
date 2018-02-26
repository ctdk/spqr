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

package state

import (
	"github.com/edsrzf/mmap-go"
	"github.com/tideland/golib/logger"
	"os"
	"time"
	"unsafe"
)

type State struct {
	createIndex  int64
	modifyIndex  int64
	lockIndex    int64
	lastIncoming time.Time
}

type Indices struct {
	CreateIndex int64
	ModifyIndex int64
	LockIndex   int64
}

func InitState(stateHolder **State, statePath string, incomingCh <-chan *Indices, errch chan<- error, doneCh chan<- struct{}) {
	fp, err := os.OpenFile(statePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		errch <- err
		return
	}
	defer fp.Close()

	logger.Debugf("Allocating %d bytes for state.", unsafe.Sizeof(**stateHolder))
	fp.Truncate(int64(unsafe.Sizeof(**stateHolder)))
	mapped, err := mmap.Map(fp, mmap.RDWR, 0)

	if err != nil {
		errch <- err
		return
	}
	defer mapped.Unmap()

	logger.Debugf("Mapped state file")
	*stateHolder = (*State)(unsafe.Pointer(&mapped[0]))

	// wait to send error back until state map is mapped
	errch <- nil
	close(errch)

	logger.Debugf("waiting for incoming events")
	for idx := range incomingCh {
		(*stateHolder).processIncomingData(idx)
	}

	doneCh <- struct{}{}
}

func (s *State) processIncomingData(idx *Indices) {
	if idx == nil {
		logger.Debugf("nil idx struct received, bailing")
		return
	}
	ut := time.Now()
	logger.Debugf("Updating state, create: %d modify: %d lock: %d at %s", idx.CreateIndex, idx.ModifyIndex, idx.LockIndex, ut)
	s.createIndex = idx.CreateIndex
	s.modifyIndex = idx.ModifyIndex
	s.lockIndex = idx.LockIndex
	s.lastIncoming = ut

	return
}

func (s *State) DoProcessIncoming(cidx int64, midx int64) bool {
	// this *may* need adjusting, in case this check ends up making it so
	// all incoming work gets skipped
	if s.modifyIndex >= midx {
		logger.Debugf("Not processing incoming payload: create: %d vs %d, modify %d vs %d", s.createIndex, cidx, s.modifyIndex, midx)
		return false
	}
	logger.Debugf("Will process payload: create: %d vs %d, modify %d vs %d", s.createIndex, cidx, s.modifyIndex, midx)
	return true
}

func (s *State) LastCreateIndex() int64 {
	return s.createIndex
}

func (s *State) LastModifyIndex() int64 {
	return s.modifyIndex
}

func (s *State) LastLockIndex() int64 {
	return s.lockIndex
}
