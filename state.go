package main

import (
	"sync"
)

const (
	MirrorStateUnknown MirrorState = iota
	MirrorStateProcessing
	MirrorStateSuccess
	MirrorStateError
)

type (
	MirrorState int

	MirrorStates struct {
		lockWrite sync.Mutex
		data      map[string]MirrorState
	}
)

func NewMirrorStates() *MirrorStates {
	states := &MirrorStates{}
	states.data = make(map[string]MirrorState)
	return states
}

func (state MirrorState) String() string {
	switch state {
	case MirrorStateProcessing:
		return "processing"

	case MirrorStateSuccess:
		return "success"

	case MirrorStateError:
		return "error"

	default:
		return "unknown"
	}
}

func (states *MirrorStates) GetState(mirror string) MirrorState {
	state, ok := states.data[mirror]
	if ok {
		return state
	}

	return MirrorStateUnknown
}

func (states *MirrorStates) SetState(mirror string, state MirrorState) {
	states.lockWrite.Lock()
	defer states.lockWrite.Unlock()

	states.data[mirror] = state
}
