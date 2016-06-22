package main

import (
	"net/http"
	"sync"
)

// ServersUpstream is representation of mirror slave set.
type ServersUpstream []ServerFollowerServer

// NewServersUpstream creates a set of mirror slaves using specified slave
// server addresses.
func NewServersUpstream(hosts []string) ServersUpstream {
	upstream := ServersUpstream{}
	for _, host := range hosts {
		upstream = append(upstream, ServerFollowerServer(host))
	}

	return upstream
}

// Propagate starts and wait all workers, which propagates requests to
// sould slave servers.
func (upstream ServersUpstream) Propagate(
	httpResource *http.Client, request PropagatableRequest,
) (successes ServersResponses, errors ServersResponses) {
	var (
		workersPropagate = sync.WaitGroup{}
		workersReceive   = sync.WaitGroup{}

		responsesError   = make(chan *ServerResponse)
		responsesSuccess = make(chan *ServerResponse)
	)

	for _, slave := range upstream {
		workersPropagate.Add(1)

		go func(slave ServerFollowerServer) {
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
