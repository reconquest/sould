package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

func feedSlaves(
	slaves []string, httpClient *http.Client,
	mirrorName string, mirrorOrigin string,
) (fedSlaves []string, errors []error) {
	var (
		work = sync.WaitGroup{}

		pipeErrors    = make(chan error)
		pipeFedSlaves = make(chan string)
	)

	for _, slave := range slaves {
		go func(slave string) {
			defer work.Done()

			// mirrorName and mirrorOrigin will be availabled
			// by link
			log.Printf("run for slave: %#v", slave)
			err := feedSlave(slave, httpClient, mirrorName, mirrorOrigin)
			if err != nil {
				log.Printf("go err: %#v", err)
				pipeErrors <- err
			} else {
				pipeFedSlaves <- slave
			}

		}(slave)

		work.Add(1)
	}

	go func() {
		for err := range pipeErrors {
			log.Printf("got err err: %#v", err)
			errors = append(errors, err)
		}
	}()

	go func() {
		for fedSlave := range pipeFedSlaves {
			log.Printf("got fedSlave fedSlave: %#v", fedSlave)
			fedSlaves = append(fedSlaves, fedSlave)
		}
	}()

	work.Wait()

	return fedSlaves, errors
}

func feedSlave(
	slave string, httpClient *http.Client,
	mirrorName string, mirrorOrigin string,
) error {
	payload := url.Values{
		"name":   {mirrorName},
		"origin": {mirrorOrigin},
	}

	response, err := httpClient.PostForm(
		"http://"+slave+"/", payload,
	)

	if err != nil {
		return fmt.Errorf(
			"can't forward request to slave '%s': %s",
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
			err = fmt.Errorf("%s, response body is empty", statusError)
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
