package main

const (
	MirrorStateUnknown MirrorState = iota
	MirrorStateSuccess
	MirrorStateFailed
)

type (
	MirrorState int

	MirrorStateTable map[string]MirrorState
)

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

func (table MirrorStateTable) GetState(mirror string) MirrorState {
	state, ok := table[mirror]
	if ok {
		return state
	}

	return MirrorStateUnknown
}

func (table MirrorStateTable) SetState(mirror string, state MirrorState) {
	table[mirror] = state
}
