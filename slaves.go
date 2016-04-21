package main

import (
	"io/ioutil"
	"net/http"

	"github.com/ajg/form"
)

// MirrorSlave is representation of slave sould server.
type MirrorSlave string

// Pull creates and sends HTTP request basing on given PullRequest variable to
// given slave server using given http client.
func (slave MirrorSlave) Pull(
	request PullRequest,
	httpResource *http.Client,
) *MirrorSlaveError {
	payload, err := form.EncodeToValues(request)
	if err != nil {
		return &MirrorSlaveError{
			Slave:        slave,
			ErrorRequest: NewError(err, "can't create payload"),
		}
	}

	response, err := httpResource.PostForm(
		"http://"+string(slave)+"/", payload,
	)
	if err != nil {
		return &MirrorSlaveError{
			Slave:        slave,
			ErrorRequest: err,
		}
	}

	if response.StatusCode == 200 || response.StatusCode == 201 {
		return nil
	}

	body, err := ioutil.ReadAll(response.Body)

	return &MirrorSlaveError{
		Slave:        slave,
		Status:       response.Status,
		StatusCode:   response.StatusCode,
		HeaderXError: response.Header.Get("X-Error"),
		ResponseBody: string(body),
		ErrorReceive: err,
	}
}
