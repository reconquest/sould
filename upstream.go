package main

import (
	"net/http"
	"sync"
)

type MirrorUpstream []MirrorSlave

func NewMirrorUpstream(hosts []string) MirrorUpstream {
	slaves := MirrorUpstream{}
	for _, host := range hosts {
		slaves = append(slaves, MirrorSlave(host))
	}

	return slaves
}

func (slaves MirrorUpstream) GetHosts() []string {
	hosts := []string{}
	for _, slave := range slaves {
		hosts = append(hosts, string(slave))
	}

	return hosts
}

// Pull runs and wait workers for all slaves, which do HTTP POST requests
// returns slice of successfully updated slaves and slice of errors, which
// can arise via running Pull() for every slave.
func (slaves MirrorUpstream) Pull(
	mirrorName string, mirrorOrigin string,
	httpClient *http.Client,
) (successMirrorUpstream MirrorUpstream, errors []error) {
	var (
		workers = sync.WaitGroup{}

		pipeErrors  = make(chan error)
		pipeUpdates = make(chan MirrorSlave)
	)

	for _, slave := range slaves {
		go func(slave MirrorSlave) {
			defer workers.Done()

			// mirrorName, mirrorOrigin and httpClient will be availabled there
			// by link
			err := slave.Pull(mirrorName, mirrorOrigin, httpClient)
			if err != nil {
				pipeErrors <- err
			} else {
				pipeUpdates <- slave
			}
		}(slave)

		workers.Add(1)
	}

	go func() {
		for err := range pipeErrors {
			errors = append(errors, err)
		}
	}()

	go func() {
		for slave := range pipeUpdates {
			successMirrorUpstream = append(successMirrorUpstream, slave)
		}
	}()

	workers.Wait()

	return successMirrorUpstream, errors
}
