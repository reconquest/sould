package main

import (
	"net/http"
	"sync"
)

// MirrorUpstream is representation of mirror slave set.
type MirrorUpstream []MirrorSlave

// NewMirrorUpstream creates a set of mirror slaves using specified slave
// server addresses.
func NewMirrorUpstream(hosts []string) MirrorUpstream {
	upstream := MirrorUpstream{}
	for _, host := range hosts {
		upstream = append(upstream, MirrorSlave(host))
	}

	return upstream
}

// PropagateRequest starts and wait all workers, which propagates requests to
// sould slave servers.
func (upstream MirrorUpstream) Propagate(
	httpResource *http.Client, request PropagatableRequest,
) (successes MirrorSlavesResponses, errors MirrorSlavesResponses) {
	var (
		workersPropagate = sync.WaitGroup{}
		workersReceive   = sync.WaitGroup{}

		responsesError   = make(chan *MirrorSlaveResponse)
		responsesSuccess = make(chan *MirrorSlaveResponse)
	)

	for _, slave := range upstream {
		workersPropagate.Add(1)

		go func(slave MirrorSlave) {
			defer workersPropagate.Done()

			response := slave.ExecuteRequest(request, httpResource)
			if response.IsSuccess() {
				responsesSuccess <- response
			} else {
				responsesError <- response
			}
		}(slave)
	}

	workersReceive.Add(1)
	go func() {
		for err := range responsesError {
			errors = append(errors, err)
		}

		workersReceive.Done()
	}()

	workersReceive.Add(1)
	go func() {
		for slave := range responsesSuccess {
			successes = append(successes, slave)
		}
		workersReceive.Done()
	}()

	workersPropagate.Wait()

	close(responsesError)
	close(responsesSuccess)

	workersReceive.Wait()

	return successes, errors
}
