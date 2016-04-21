package main

import (
	"errors"

	"github.com/seletskiy/hierr"
)

// MirrorSlaveError represents information about occurred error while working
// with sould slave server.
//
// Also, MirrorSlaveError implements Error interfaces.
type MirrorSlaveError struct {
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

	// ErrorRequest is error which has been occurred in http communication
	// session with slave server.
	ErrorRequest error

	// ErrorReceive is error which has been occurred while receiving response
	// from slave server
	ErrorReceive error
}

func (err MirrorSlaveError) Error() string {
	if err.ErrorRequest != nil {
		return err.ErrorRequest.Error()
	}

	if len(err.HeaderXError) > 0 {
		return "received response with status " + err.Status +
			": " + err.HeaderXError
	}

	message := "received unexpected and ambigious response with status " +
		err.Status + ", without X-Error header"

	if err.ErrorReceive != nil {
		message += ", and an error occurred while " +
			"receiving response: " + err.ErrorReceive.Error()
	}

	message += ", received response body"
	if len(err.ResponseBody) > 0 {
		message += ": " + err.ResponseBody
	} else {
		message += " is empty"
	}

	return message
}

func (err MirrorSlaveError) HierarchicalError() string {
	if err.ErrorRequest != nil {
		return err.ErrorRequest.Error()
	}

	hierarchical := errors.New(err.Status)

	if err.ErrorReceive != nil {
		hierarchical = hierr.Push(hierarchical, err.ErrorReceive)
	}

	if len(err.ResponseBody) > 0 {
		if len(err.HeaderXError) > 0 {
			hierarchical = hierr.Push(hierarchical, err.ResponseBody)
		} else {
			hierarchical = hierr.Push(
				hierarchical,
				"X-Error header is missing",
				hierr.Errorf(
					err.ResponseBody,
					"unexpected and ambigious response body",
				),
			)
		}

		return hierarchical.Error()
	}

	hierarchical = hierr.Push(
		hierarchical,
		hierr.Errorf(
			err.ResponseBody,
			"unexpected and ambigious response without body",
		),
	)

	if len(err.HeaderXError) > 0 {
		hierarchical = hierr.Push(
			hierarchical,
			err.HeaderXError,
		)
	} else {
		hierarchical = hierr.Push(
			hierarchical,
			"X-Error header is missing",
		)
	}

	return hierarchical.Error()
}
