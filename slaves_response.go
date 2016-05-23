package main

import (
	"errors"
	"fmt"

	"github.com/seletskiy/hierr"
)

type MirrorSlavesResponses []*MirrorSlaveResponse

// GetHosts of given mirror slave servers.
func (responses MirrorSlavesResponses) GetHosts() []string {
	hosts := []string{}
	for _, response := range responses {
		hosts = append(hosts, string(response.Slave))
	}

	return hosts
}

// MirrorSlaveResponse represents information about result of request
// propagation to sould slave server.
//
// Also, MirrorSlaveResponse implements Error interfaces.
type MirrorSlaveResponse struct {
	// Slave is problematic mirror slave server.
	Slave MirrorSlave

	// Status is HTTP status which has been received from slave server response.
	Status string

	// StatusCode is same as Status but numeric.
	StatusCode int

	// ResponseBody is contents of response body which has been received from
	// slave server response.
	ResponseBody string

	// HeaderXError is string representation of error which has been occurred
	// on slave server.
	//
	// Actually, it's X-Error HTTP Header which has been received from slave
	// server response.
	HeaderXError string

	HeaderXSuccess string

	// ErrorRequest is error which has been occurred in http communication
	// session with slave server.
	ErrorRequest error

	// ErrorReceive is error which has been occurred while receiving response
	// from slave server
	ErrorReceive error
}

// Error returns plain one-line string representation of occurred error, this
// method should be used for saving error to sould error logs.
func (response MirrorSlaveResponse) Error() string {
	if response.ErrorRequest != nil {
		return response.ErrorRequest.Error()
	}

	if len(response.HeaderXError) > 0 {
		return "received response with status " + response.Status +
			": " + response.HeaderXError
	}

	message := "received unexpected and ambigious response with status " +
		response.Status + ", without X-Error header"

	if response.ErrorReceive != nil {
		message += ", and an error occurred while " +
			"receiving response: " + response.ErrorReceive.Error()
	}

	message += ", received response body"
	if len(response.ResponseBody) > 0 {
		message += ": " + response.ResponseBody
	} else {
		message += " is empty"
	}

	return message
}

// HierarchicalError returns hierarchical (with unicode symbols) string
// representation of occurred error, this method used by hierr package for
// sending occurred slave errors to user as part of http response.
func (response MirrorSlaveResponse) HierarchicalError() string {
	if response.ErrorRequest != nil {
		return response.ErrorRequest.Error()
	}

	hierarchical := errors.New(response.Status)

	if response.ErrorReceive != nil {
		hierarchical = hierr.Push(hierarchical, response.ErrorReceive)
	}

	if len(response.ResponseBody) > 0 {
		if len(response.HeaderXError) > 0 {
			hierarchical = hierr.Push(hierarchical, response.ResponseBody)
		} else {
			hierarchical = hierr.Push(
				hierarchical,
				"X-Error header is missing",
				hierr.Errorf(
					response.ResponseBody,
					"unexpected and ambigious response body",
				),
			)
		}

		return hierarchical.Error()
	}

	hierarchical = hierr.Push(
		hierarchical,
		hierr.Errorf(
			response.ResponseBody,
			"unexpected and ambigious response without body",
		),
	)

	if len(response.HeaderXError) > 0 {
		hierarchical = hierr.Push(
			hierarchical,
			response.HeaderXError,
		)
	} else {
		hierarchical = hierr.Push(
			hierarchical,
			"X-Error header is missing",
		)
	}

	return hierarchical.Error()
}

func (response MirrorSlaveResponse) IsSuccess() bool {
	fmt.Printf("XXXXXX slaves_response.go:141: response.HeaderXSuccess: %#v\n", response.HeaderXSuccess)
	return response.HeaderXSuccess == "true"
}
