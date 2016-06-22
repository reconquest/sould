package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/seletskiy/hierr"
)

// HandlePullRequest handles new request for pulling changeset, starts request
// propagation to slave servers in parallel mode (if this server is master),
// pulls changeset and waits for propagation process.
func (server *ServerHTTP) HandlePullRequest(
	response http.ResponseWriter, request *PullRequest,
) {
	var propagation *RequestPropagation
	if server.IsMaster() {
		propagation = server.propagatePullRequest(request)
	}

	changesetPulled, pullError := server.ServePullRequest(request)
	if pullError != nil {
		logger.Errorf(
			"can't pull mirror %s (%s) changeset: %s",
			request.MirrorName, request.MirrorOrigin,
			pullError,
		)
	}

	if server.IsSlave() {
		if !changesetPulled {
			response.Header().Set("X-Error", oneLineError(pullError))

			http.Error(
				response,
				hierr.Errorf(pullError, "pull changeset failed").Error(),
				http.StatusInternalServerError,
			)
			return
		}

		response.Header().Set("X-Success", "true")
		return
	}

	propagation.Wait()

	var status int
	switch {
	case propagation.IsAllSlavesFailed():
		status = http.StatusServiceUnavailable
	case propagation.IsAnySlaveFailed():
		status = http.StatusBadGateway
	case !changesetPulled:
		status = http.StatusInternalServerError
	default:
		response.WriteHeader(http.StatusOK)
		response.Header().Set("X-Success", "true")
		return
	}

	err := errors.New(request.MirrorName + " (" + request.MirrorOrigin + ")")

	if !changesetPulled {
		err = hierr.Push(
			err,
			hierr.Push(
				"master",
				hierr.Errorf(pullError, "can't pull mirror changeset"),
			),
		)
	}

	for _, slaveResponse := range propagation.ResponsesError() {
		err = hierr.Push(
			err,
			hierr.Errorf(slaveResponse, "slave "+string(slaveResponse.Slave)),
		)
	}

	response.Header().Set("X-Error", request.MirrorName)
	http.Error(response, err.Error(), status)
}

// ServePullRequest exactly serves request for pulling changeset, pulls and
// spoofs changeset.
func (server *ServerHTTP) ServePullRequest(
	request *PullRequest,
) (bool, error) {
	mirror, created, err := server.GetMirror(
		request.MirrorName, request.MirrorOrigin,
	)
	if err != nil {
		return false, NewError(err, "can't obtain mirror")
	}

	if created {
		logger.Infof("mirror %s successfully created", mirror.String())
	}

	logger.Infof("fetching mirror %s changeset", mirror.String())

	server.states.Set(request.MirrorName, MirrorStateProcessing)

	err = mirror.Fetch()
	if err != nil {
		server.states.Set(request.MirrorName, MirrorStateError)
		return false, NewError(err, "can't fetch changeset")
	}

	logger.Infof(
		"mirror %s (%s) changeset fetched",
		request.MirrorName, request.MirrorOrigin,
	)

	if request.Spoof {
		logger.Infof(
			"spoofing mirror %s changeset %s -> %s",
			mirror.String(), request.SpoofingBranch, request.SpoofingTag,
		)

		err = mirror.SpoofChangeset(request.SpoofingBranch, request.SpoofingTag)
		if err != nil {
			server.states.Set(request.MirrorName, MirrorStateError)
			return false, NewError(
				err,
				"can't spoof mirror changeset %s -> %s",
				request.SpoofingBranch, request.SpoofingTag,
			)
		}

		logger.Infof(
			"mirror %s changeset spoofed %s -> %s",
			mirror.String(), request.SpoofingBranch, request.SpoofingTag,
		)
	}

	server.states.Set(request.MirrorName, MirrorStateSuccess)

	return true, nil
}

// propagatePullRequest propagates specified PullRequest to mirror upstream,
// waits for result of propagation and logs results and errors.
// Returns instance of running propagation operation.
func (server *ServerHTTP) propagatePullRequest(
	request *PullRequest,
) *RequestPropagation {
	var (
		mirrors = server.GetMirrorUpstream()

		propagation = NewRequestPropagation(
			server.httpResource, mirrors, request,
		)
	)

	logger.Infof(
		"propagating pull request for mirror %s (%s)",
		request.MirrorName, request.MirrorOrigin,
	)

	propagation.Start()

	go func() {
		propagation.Wait()

		logPropagation(
			fmt.Sprintf(
				"pull %s (%s)", request.MirrorName, request.MirrorOrigin,
			),
			propagation,
		)
	}()

	return propagation
}
