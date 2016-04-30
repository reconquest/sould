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

// GetHosts of given mirror slave servers.
func (upstream MirrorUpstream) GetHosts() []string {
	hosts := []string{}
	for _, slave := range upstream {
		hosts = append(hosts, string(slave))
	}

	return hosts
}

// PropagatePullRequest starts and wait all workers, which propagates POST
// requests to sould slave servers.
// Returns slice of successfully updated slaves and slice of slave errors,
// which can arise via running Pull() for every slave.
func (upstream MirrorUpstream) PropagatePullRequest(
	httpResource *http.Client, request PullRequest,
) (successes []MirrorSlave, errors []*MirrorSlaveError) {
	var (
		workersPull     = sync.WaitGroup{}
		workersResponse = sync.WaitGroup{}

		responsesError   = make(chan *MirrorSlaveError)
		responsesSuccess = make(chan MirrorSlave)
	)

	for _, slave := range upstream {
		workersPull.Add(1)

		go func(slave MirrorSlave) {
			defer workersPull.Done()

			err := slave.Pull(request, httpResource)
			if err != nil {
				responsesError <- err
			} else {
				responsesSuccess <- slave
			}
		}(slave)
	}

	workersResponse.Add(1)
	go func() {
		for err := range responsesError {
			errors = append(errors, err)
		}

		workersResponse.Done()
	}()

	workersResponse.Add(1)
	go func() {
		for slave := range responsesSuccess {
			successes = append(successes, slave)
		}
		workersResponse.Done()
	}()

	workersPull.Wait()

	close(responsesError)
	close(responsesSuccess)

	workersResponse.Wait()

	return successes, errors
}
