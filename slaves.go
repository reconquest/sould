package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type MirrorSlave string
type MirrorSlaves []MirrorSlave

func NewMirrorSlaves(hosts []string) MirrorSlaves {
	slaves := MirrorSlaves{}
	for _, host := range hosts {
		slaves = append(slaves, MirrorSlave(host))
	}

	return slaves
}

func (slaves MirrorSlaves) GetHosts() []string {
	hosts := []string{}
	for _, slave := range slaves {
		hosts = append(hosts, string(slave))
	}

	return hosts
}

// 'MirrorSlaves.Pull()' runs and wait workers for all slaves, which do HTTP
// POST requests
// returns slice of successfully updated slaves and slice of errors, which
// can arise via running Pull() for every slave.
func (slaves MirrorSlaves) Pull(
	mirrorName string, mirrorOrigin string,
	httpClient *http.Client,
) (updatedSlaves MirrorSlaves, errors []error) {

	var (
		workers = sync.WaitGroup{}

		pipeErrors  = make(chan error)
		pipeUpdates = make(chan MirrorSlave)
	)

	for _, slave := range slaves {
		go func(slave MirrorSlave) {
			defer workers.Done()

			// mirrorName and mirrorOrigin will be availabled by link
			log.Printf("run for slave: %#v", slave)
			err := slave.Pull(mirrorName, mirrorOrigin, httpClient)
			if err != nil {
				log.Printf("go err: %#v", err)
				pipeErrors <- err
			} else {
				pipeUpdates <- slave
			}

		}(slave)

		workers.Add(1)
	}

	go func() {
		for err := range pipeErrors {
			log.Printf("got err err: %#v", err)
			errors = append(errors, err)
		}
	}()

	go func() {
		for slave := range pipeUpdates {
			log.Printf("got updatedSlave updatedSlave: %#v", slave)
			updatedSlaves = append(updatedSlaves, slave)
		}
	}()

	workers.Wait()

	return updatedSlaves, errors
}

// 'MirrorSlave.Pull()' do HTTP POST request to slave with mirror name and
// origin
func (slave MirrorSlave) Pull(
	mirrorName string, mirrorOrigin string,
	httpClient *http.Client,
) error {
	payload := url.Values{
		"name":   {mirrorName},
		"origin": {mirrorOrigin},
	}

	response, err := httpClient.PostForm(
		"http://"+string(slave)+"/", payload,
	)

	if err != nil {
		return fmt.Errorf(
			"can't propagate request to slave '%s': %s",
			slave, err.Error(),
		)
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		statusError := fmt.Errorf(
			"slave '%s' was unwell, http status is '%s'",
			slave, response.Status,
		)

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf(
				"%s, can't read response body: %s",
				statusError, err,
			)
		}

		defer response.Body.Close()

		if string(body) == "" {
			err = fmt.Errorf(
				"%s, response body is empty",
				statusError,
			)

		} else {
			err = fmt.Errorf(
				"%s, response body: %s",
				statusError, string(body),
			)
		}

		return err
	}

	return nil
}
