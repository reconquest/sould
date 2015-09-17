package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type MirrorSlave string

// Pull do HTTP POST request to slave with mirror name and
// origin
func (slave MirrorSlave) Pull(
	request RequestPull,
	httpClient *http.Client,
) error {
	payload := url.Values{
		"name":   {request.MirrorName},
		"origin": {request.MirrorOrigin},
	}

	if request.ShouldSpoof {
		payload.Set("spoof", "true")
		payload.Set("branch", request.SpoofBranch)
		payload.Set("tag", request.SpoofTag)
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
				"%s, response body is empty", statusError,
			)

		} else {
			err = fmt.Errorf(
				"%s, response body: %s", statusError, string(body),
			)
		}

		return err
	}

	return nil
}
