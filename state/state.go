package state

import (
	"github.com/edsrzf/mmap-go"
	"os"
	"time"
	"unsafe"
)

type State struct {
	createIndex int
	modifyIndex int
	lockIndex int
	lastIncoming time.Time
}

type Indices struct {
	CreateIndex int
	ModifyIndex int
	LockIndex int
}

func InitState(stateHolder *State, statePath string, incomingCh <-chan *Indices, errch chan<- error, doProcess chan<- bool, doneCh chan<- struct{}) {
	fp, err := os.OpenFile(statePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		errch <- err
		return
	}
	defer fp.Close()
	
	fp.Truncate(int64(unsafe.Sizeof(stateHolder)))
	mapped, err := mmap.Map(fp, mmap.RDWR, 0)

	if err != nil {
		errch <- err
		return
	}
	defer mapped.Unmap()

	errch <- nil
	close(errch)

	stateHolder = (*State)(unsafe.Pointer(&mapped[0]))

	for idx := range incomingCh {
		stateHolder.processIncomingData(idx, doProcess)
	}
	doneCh <- struct{}{}
}

func (s *State) processIncomingData(idx *Indices, doProcess chan<- bool) {
	// this *may* need adjusting, in case this check ends up making it so
	// all incoming work gets skipped
	if s.createIndex >= idx.CreateIndex && s.modifyIndex >= idx.ModifyIndex {
		doProcess <- false
		return
	}
	s.createIndex = idx.CreateIndex
	s.modifyIndex = idx.ModifyIndex
	s.lockIndex = idx.LockIndex
	s.lastIncoming = time.Now()

	doProcess <- true
	return
}

func (s *State) LastCreateIndex() int {
	return s.createIndex
}

func (s *State) LastModifyIndex() int {
	return s.modifyIndex
}

func (s *State) LastLockIndex() int {
	return s.lockIndex
}
