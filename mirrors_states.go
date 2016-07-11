package main

import (
	"sync"
)

const (
	// MirrorStateUnknown is default mirror state
	MirrorStateUnknown MirrorState = iota

	// MirrorStateProcessing is used if mirror fetches remote data now.
	MirrorStateProcessing

	// MirrorStateSuccess is used if mirror has been pulled.
	MirrorStateSuccess

	// MirrorStateError is used if an error occurred during mirror pull.
	MirrorStateError
)

type (
	// MirrorState is representation of current state of specified mirror.
	MirrorState int

	// MirrorStates is thread-safe table for storing mirror states in memory.
	MirrorStates struct {
		sync.Mutex
		data map[string]MirrorState
	}
)

// String returns string representation of mirror's state.
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

// NewMirrorStates creates a new empty mirror states table.
func NewMirrorStates() *MirrorStates {
	states := &MirrorStates{
		data: map[string]MirrorState{},
	}
	return states
}

// Get state of specified mirror.
func (states *MirrorStates) Get(mirror string) MirrorState {
	state, ok := states.data[mirror]
	if ok {
		return state
	}

	return MirrorStateUnknown
}

// Set stores information about specified mirror and state.
func (states *MirrorStates) Set(mirror string, state MirrorState) {
	states.Lock()
	defer states.Unlock()

	states.data[mirror] = state
}
