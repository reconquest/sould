package main

import (
	"net/http"
	"sync"
)

// RequestPropagation is representation of propagation pull (and spoof) or
// status requests basing on given request to specified slaves mirror
// upstream.
type RequestPropagation struct {
	worker *sync.WaitGroup

	upstream MirrorUpstream
	request  PropagatableRequest
	resource *http.Client

	successes MirrorSlavesResponses
	errors    MirrorSlavesResponses
}

// NewRequestPropagation returns reference to operation of propagation
// http request basing on given request and using given http client and mirror
// upstream.
func NewRequestPropagation(
	httpResource *http.Client,
	upstream MirrorUpstream,
	request PropagatableRequest,
) *RequestPropagation {
	return &RequestPropagation{
		worker:   &sync.WaitGroup{},
		resource: httpResource,
		request:  request,

		upstream: upstream,
	}
}

// Start starts operation of propagation request and locks internal
// waiting group which will be unlocked when operation of propagation is done.
func (propagation *RequestPropagation) Start() {
	propagation.worker.Add(1)

	go propagation.propagate()
}

// Wait blocks until the operation of propagation is in process.
func (propagation *RequestPropagation) Wait() {
	propagation.worker.Wait()
}

func (propagation *RequestPropagation) propagate() {
	defer propagation.worker.Done()

	successes, errors := propagation.upstream.Propagate(
		propagation.resource, propagation.request,
	)

	propagation.successes = successes
	propagation.errors = errors
}

// ResponsesSuccess returns slice of complex structured responses from success
// slaves.
func (propagation RequestPropagation) ResponsesSuccess() MirrorSlavesResponses {
	return propagation.successes
}

// ResponsesError returns slice of complex structured error responses from
// problematic slaves.
func (propagation RequestPropagation) ResponsesError() MirrorSlavesResponses {
	return propagation.errors
}

// IsAllSlavesFailed returns true only if given slaves upstream is not empty
// and all slaves from given upstream are failed.
func (propagation *RequestPropagation) IsAllSlavesFailed() bool {
	return len(propagation.upstream) > 0 &&
		len(propagation.errors) == len(propagation.upstream)
}

// IsAnySlaveFailed returns true if propagation have at least one failed slave.
func (propagation *RequestPropagation) IsAnySlaveFailed() bool {
	return len(propagation.errors) > 0
}
