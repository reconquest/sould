package main

import (
	"net/http"

	"github.com/pquerna/ffjson/ffjson"
)

const (
	// ServerStatusResponseIndentation uses for json and toml marshalers.
	ServerStatusResponseIndentation = "    "
)

// HandleStatusRequest handles requests for sould mirrors status.
func (server *Server) HandleStatusRequest(
	response http.ResponseWriter, request StatusRequest,
) {
	status := server.serveStatusRequest(request)

	var err error
	var buffer []byte
	switch {
	case request.IsFormatJSON():
		buffer, err = status.MarshalJSON()
		if err != nil {
			err = NewError(err, "can't encode json")
			break
		}

	case request.IsFormatTOML():
		buffer, err = status.MarshalTOML()
		if err != nil {
			err = NewError(err, "can't encode toml")
			break
		}

	default:
		buffer = status.MarshalHierarchical()
	}

	if err != nil {
		logger.Errorf("%s, status: %#v", err, status)
		response.Header().Set("X-Error", err.Error())
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Header().Set("X-Success", "true")
	response.WriteHeader(http.StatusOK)
	response.Write(buffer)
}

func (server *Server) serveStatusRequest(
	request StatusRequest,
) ServerStatus {
	var propagation *RequestPropagation
	if server.IsMaster() {
		propagation = server.propagateStatusRequest(request)
	}

	mirrors, errors := server.getMirrorsStatuses()

	status := ServerStatus{
		BasicServerStatus: BasicServerStatus{
			Role:    server.GetRole(),
			Mirrors: mirrors,
			Total:   len(mirrors),
		},
	}

	for _, err := range errors {
		if status.Error == "" {
			status.Error = err.Error()
			status.HierarchicalError = err.HierarchicalError()
		}

		logger.Error(err)
	}

	if server.IsSlave() {
		status.Role = "slave"
		return status
	}

	propagation.Wait()

	status.Upstream = getUpstreamStatus(propagation)

	return status
}

func (server *Server) getMirrorsStatuses() ([]MirrorStatus, []Error) {
	var statuses []MirrorStatus
	var errors []Error

	mirrors, err := getAllMirrors(server.GetStorageDir())
	if err != nil {
		errors = append(errors, NewError(err, "can't get mirrors list"))
	}

	for _, mirrorName := range mirrors {
		status := MirrorStatus{
			Name:       mirrorName,
			State:      server.states.Get(mirrorName).String(),
			ModifyDate: 0,
		}

		mirror, err := GetMirror(server.GetStorageDir(), mirrorName)
		if err != nil {
			errors = append(
				errors, NewError(err, "can't get mirror %s", mirrorName),
			)

			status.State = "error"
		} else {
			status.ModifyDate, err = mirror.GetModifyDateUnix()
			if err != nil {
				errors = append(
					errors, NewError(
						err,
						"can't get mirror %s modify date", mirrorName,
					),
				)

				status.State = "error"
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, errors
}

func (server *Server) propagateStatusRequest(
	request StatusRequest,
) *RequestPropagation {
	var (
		mirrors = server.GetServersUpstream()

		propagation = NewRequestPropagation(
			server.httpResource, mirrors, request,
		)
	)

	logger.Info("propagating status request")

	propagation.Start()

	go func() {
		propagation.Wait()

		logPropagation("status", propagation)
	}()

	return propagation
}

func getUpstreamStatus(propagation *RequestPropagation) UpstreamStatus {
	var (
		successes = propagation.ResponsesSuccess()
		errors    = propagation.ResponsesError()

		status = UpstreamStatus{
			Total: len(successes) + len(errors),
			Error: len(errors),
		}
	)

	for _, response := range successes {
		var slave ServerStatus

		err := ffjson.Unmarshal([]byte(response.ResponseBody), &slave)
		if err != nil {
			logger.Errorf(
				"can't decode JSON response from %s: %s, response:\n%s",
				response.Slave, err, response.ResponseBody,
			)

			err := NewError(
				err,
				"can't decode JSON response",
			)

			slave.Error = err.Error()
			slave.HierarchicalError = err.HierarchicalError()

			status.Error++
		} else {
			status.Success++
		}

		slave.Address = string(response.Slave)

		status.Slaves = append(status.Slaves, slave)
	}

	for _, response := range errors {
		status.Slaves = append(status.Slaves, ServerStatus{
			BasicServerStatus: BasicServerStatus{
				Address:           string(response.Slave),
				Error:             response.Error(),
				HierarchicalError: response.HierarchicalError(),
			},
		})
	}

	status.ErrorPercent = percent(status.Error, status.Total)
	status.SuccessPercent = percent(status.Success, status.Total)

	return status
}
