package main

import (
	"sync"
)

const (
	MirrorStateUnknown MirrorState = iota
	MirrorStateSuccess
	MirrorStateFailed
)

type (
	MirrorState int

	MirrorStateTable struct {
		lockWrite sync.Mutex
		data      map[string]MirrorState
	}
)

func NewMirrorStateTable() *MirrorStateTable {
	table := &MirrorStateTable{}
	table.data = make(map[string]MirrorState)
	return table
}

func (state MirrorState) String() string {
	switch state {
	case MirrorStateSuccess:
		return "success"

	case MirrorStateFailed:
		return "failed"

	default:
		return "unknown"
	}
}

func (table *MirrorStateTable) GetState(mirror string) MirrorState {
	state, ok := table.data[mirror]
	if ok {
		return state
	}

	return MirrorStateUnknown
}

func (table *MirrorStateTable) SetState(mirror string, state MirrorState) {
	table.lockWrite.Lock()
	defer table.lockWrite.Unlock()

	table.data[mirror] = state
}
