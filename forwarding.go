package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

func feedSlaves(
	slaves []string,
	httpClient *http.Client,
	mirrorName string,
	mirrorCloneURL string,
) ([]string, []error) {
	var (
		work = sync.WaitGroup{}

		pipeErrors    = make(chan error)
		pipeFedSlaves = make(chan string)
	)

	for _, slave := range slaves {
		go func(slave string) {
			//  mirrorName and mirrorCloneURL will be availabled
			// by link
			err := feedSlave(slave, httpClient, mirrorName, mirrorCloneURL)
			if err != nil {
				pipeErrors <- err
			} else {
				pipeFedSlaves <- slave
			}

			work.Done()
		}(slave)

		work.Add(1)
	}

	var (
		fedSlaves []string
		errors    []error
	)

	select {
	case err := <-pipeErrors:
		errors = append(errors, err)

	case fedSlave := <-pipeFedSlaves:
		fedSlaves = append(fedSlaves, fedSlave)

	default:
		work.Wait()
	}

	return fedSlaves, errors
}

func feedSlave(
	slave string,
	httpClient *http.Client,
	mirrorName string,
	mirrorCloneURL string,
) error {
	payload := url.Values{
		"name": {mirrorName},
		"url":  {mirrorCloneURL},
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

	if response.StatusCode != 200 {
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
