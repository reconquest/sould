package main

import (
	"net/http"
	"os"
)

// HandleTarRequest handles requests for downloading tar archives of specified
// revision.
func (server *MirrorServer) HandleTarRequest(
	response http.ResponseWriter, request TarRequest,
) {
	mirror, err := GetMirror(server.GetStorageDir(), request.MirrorName)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warningf("mirror %s not found", request.MirrorName)
			response.WriteHeader(http.StatusNotFound)
			return
		}

		logger.Errorf("can't get mirror %s: %s", request.MirrorName, err)
		http.Error(
			response,
			"can't get mirror: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	mirrorState := server.states.Get(request.MirrorName)
	if mirrorState == MirrorStateUnknown || mirrorState == MirrorStateError {
		server.states.Set(request.MirrorName, MirrorStateProcessing)

		logger.Infof("fetching mirror %s changeset", mirror.String())

		err = mirror.Fetch()
		if err != nil {
			logger.Errorf(
				"can't fetch mirror %s changeset: %s",
				mirror.String(), err,
			)

			mirrorState = MirrorStateError
		} else {
			mirrorState = MirrorStateSuccess
		}

		server.states.Set(request.MirrorName, mirrorState)
	}

	modifyDate, err := mirror.GetModifyDate()
	if err != nil {
		logger.Infof(
			"can't get mirror %s modify date",
			request.MirrorName, err.Error(),
		)

		http.Error(
			response,
			"can't get mirror modify date: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	response.Header().Set("X-State", mirrorState.String())
	response.Header().Set("X-Date", modifyDate.UTC().Format(http.TimeFormat))
	response.Header().Set("Content-Type", "application/x-tar")

	err = mirror.Archive(response, request.Reference)
	if err != nil {
		logger.Errorf("can't archive mirror %s: %s", mirror.String(), err)

		http.Error(
			response,
			"can't archive mirror: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}
}
