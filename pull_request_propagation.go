package main

import (
	"net/http"
	"sync"
)

// PullRequestPropagation is representation of propagation pull (and spoof)
// changeset request basing on given request to specified slaves mirror
// upstream.
type PullRequestPropagation struct {
	worker *sync.WaitGroup

	upstream MirrorUpstream
	request  PullRequest
	resource *http.Client

	successes []MirrorSlave
	errors    []*MirrorSlaveError
}

// NewPullRequestPropagation returns reference to operation of propagation pull
// request  basing on given request and using given http client and mirror
// upstream.
func NewPullRequestPropagation(
	httpResource *http.Client, upstream MirrorUpstream, request PullRequest,
) *PullRequestPropagation {
	return &PullRequestPropagation{
		worker:   &sync.WaitGroup{},
		resource: httpResource,
		request:  request,

		upstream: upstream,
	}
}

// Start starts operation of propagation pull request and locks internal
// waiting group which will be unlocked when operation of propagation is done.
func (propagation *PullRequestPropagation) Start() {
	propagation.worker.Add(1)

	go propagation.propagate()
}

// Wait blocks until the operation of propagation is in process.
func (propagation *PullRequestPropagation) Wait() {
	propagation.worker.Wait()
}

func (propagation *PullRequestPropagation) propagate() {
	defer propagation.worker.Done()

	successes, errors := propagation.upstream.PropagatePullRequest(
		propagation.resource, propagation.request,
	)

	propagation.successes = successes
	propagation.errors = errors
}

// SlavesSuccess returns slice of slaves which successfully pulled changeset.
func (propagation PullRequestPropagation) SlavesSuccess() []MirrorSlave {
	return propagation.successes
}

// SlavesErrors returns slice of complex structured errors which has been
// occurred while propagation request to slaves.
func (propagation PullRequestPropagation) SlavesErrors() []*MirrorSlaveError {
	return propagation.errors
}

// IsAllSlavesFailed returns true only if given slaves upstream is not empty
// and all slaves from given upstream are failed.
func (propagation *PullRequestPropagation) IsAllSlavesFailed() bool {
	return len(propagation.upstream) > 0 &&
		len(propagation.errors) == len(propagation.upstream)
}

// IsAnySlaveFailed returns true if propagation have at least one failed slave.
func (propagation *PullRequestPropagation) IsAnySlaveFailed() bool {
	return len(propagation.errors) > 0
}
