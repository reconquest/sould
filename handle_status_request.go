package main

import (
	"encoding/json"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/seletskiy/hierr"
)

// HandleStatusRequest handles requests for sould mirrors status.
func (server *MirrorServer) HandleStatusRequest(
	response http.ResponseWriter, request StatusRequest,
) {
	status := server.serveStatusRequest(request)

	response.Header().Set("X-Success", "true")

	var err error
	switch {
	case request.FormatJSON():
		err = json.NewEncoder(response).Encode(status)
		if err != nil {
			err = NewError(err, "can't encode json")
		}
	case request.FormatTOML():
		err = toml.NewEncoder(response).Encode(status)
		if err != nil {
			err = NewError(err, "can't encode toml")
		}
	default:
		response.Write([]byte(hierr.String(status)))
	}

	if err != nil {
		response.Header().Set("X-Error", err.Error())
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Header().Set("X-Success", "true")
	response.WriteHeader(http.StatusOK)
}

func (server *MirrorServer) serveStatusRequest(
	request StatusRequest,
) interface{} {
	var propagation *RequestPropagation
	if server.IsMaster() {
		propagation = server.propagateStatusRequest(request)
	}

	mirrors, errors := server.getMirrorsStatuses()

	status := ServerStatus{
		Mirrors: mirrors,
	}

	for _, err := range errors {
		if status.Error == nil {
			status.Error = err
		}

		logger.Error(err)
	}

	if server.IsSlave() {
		return status
	}

	propagation.Wait()

	return MasterServerStatus{
		Master:   status,
		Upstream: getUpstreamStatus(propagation),
	}
}

func (server *MirrorServer) getMirrorsStatuses() ([]MirrorStatus, []error) {
	var statuses []MirrorStatus
	var errors []error

	mirrors, err := getAllMirrors(server.GetStorageDir())
	if err != nil {
		errors = append(errors, NewError(err, "can't get mirrors list"))
	}

	for _, mirrorName := range mirrors {
		mirror, err := GetMirror(server.GetStorageDir(), mirrorName)
		if err != nil {
			errors = append(
				errors, NewError(err, "can't get mirror %s"),
			)
			continue
		}

		modifyDate, err := mirror.GetModifyDate()
		if err != nil {
			errors = append(
				errors, NewError(err, "can't get mirror %s modify date"),
			)
		}

		status := MirrorStatus{
			Name:       mirror.Name,
			State:      server.states.Get(mirror.Name),
			ModifyDate: modifyDate.Unix(),
		}

		statuses = append(statuses, status)
	}

	return statuses, errors
}

func (server *MirrorServer) propagateStatusRequest(
	request StatusRequest,
) *RequestPropagation {
	var (
		mirrors = server.GetMirrorUpstream()

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

		_, err := toml.Decode(response.ResponseBody, &slave)
		if err != nil {
			status.Error++

			slave.Error = NewError(err, "can't decode toml response")
		} else {
			status.Success++
		}

		slave.Address = string(response.Slave)

		status.Slaves = append(status.Slaves, slave)
	}

	for _, response := range errors {
		status.Slaves = append(status.Slaves, ServerStatus{
			Address: string(response.Slave),
			Error:   response,
		})
	}

	status.ErrorPercent = percent(status.Error, status.Total)
	status.SuccessPercent = percent(status.Success, status.Total)

	return status
}
