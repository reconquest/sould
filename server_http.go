package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func (server *MirrorServer) ListenHTTP() error {
	http.Handle("/", server)

	err := http.ListenAndServe(server.GetListenAddress(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (server *MirrorServer) ServeHTTP(
	response http.ResponseWriter, request *http.Request,
) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()

	method := strings.ToUpper(request.Method)

	switch method {
	case "POST":
		server.HandlePOST(response, request)

	case "GET":
		server.HandleGET(response, request)

	default:
		response.WriteHeader(http.StatusMethodNotAllowed)

		log.Printf("got request with unsupported method: '%s'", method)
	}
}

func (server *MirrorServer) HandlePOST(
	response http.ResponseWriter, request *http.Request,
) {
	err := request.ParseForm()
	if err != nil {
		log.Println(err)
		writeResponseError(response, http.StatusBadRequest, err.Error())

		return
	}

	var (
		mirrorName   = request.FormValue("name")
		mirrorOrigin = request.FormValue("origin")
	)

	for paramName, paramValue := range map[string]string{
		"name":   mirrorName,
		"origin": mirrorOrigin,
	} {
		if paramValue == "" {
			err = fmt.Errorf(
				"'%s' param not found in request", paramName,
			)

			log.Printf("%s, request = '%#v'", err.Error(), request.Form)

			writeResponseError(
				response, http.StatusBadRequest,
				err.Error(),
			)

			return
		}
	}

	log.Printf(
		"got pull request for mirror with name = %s, origin = %s",
		mirrorName, mirrorOrigin,
	)

	mirror, created, err := server.GetMirror(mirrorName, mirrorOrigin)

	if err != nil {
		log.Println(err)
		writeResponseError(
			response, http.StatusInternalServerError, err.Error(),
		)

		return
	}

	if created {
		log.Printf(
			"mirror '%s' successfully created",
			mirrorName,
		)
	}

	var (
		responseMessages []string
		httpStatus       int
		pullFailed       bool
	)

	err = mirror.Pull()
	if err != nil {
		server.stateTable.SetState(mirrorName, MirrorStateFailed)

		httpStatus = http.StatusInternalServerError
		pullFailed = true

		err = fmt.Errorf(
			"can't update mirror '%s': %s",
			mirrorName, err.Error(),
		)

		log.Println(err)

		// an error occurred during mirror pull, master sould should be
		// tolerant and try to forward request to slaves
		if server.IsMaster() {
			responseMessages = append(responseMessages, err.Error())
		} else {
			writeResponseError(response, httpStatus, err.Error())

			return
		}
	} else {
		server.stateTable.SetState(mirrorName, MirrorStateSuccess)
		httpStatus = http.StatusOK

		log.Printf("mirror '%s' successfully updated", mirrorName)
	}

	if server.IsMaster() {
		slaves := server.GetSlaves()
		if len(slaves) > 0 {
			fedSlaves, errors := feedSlaves(
				slaves, server.httpClient,
				mirrorName, mirrorOrigin,
			)

			if len(fedSlaves) > 0 {
				log.Printf(
					"request successfully forwarded to slaves %s",
					strings.Join(fedSlaves, ", "),
				)
			}

			if len(errors) > 0 {
				httpStatus = http.StatusBadGateway

				for _, err := range errors {
					log.Println(err)

					responseMessages = append(responseMessages, err.Error())
				}

				if len(fedSlaves) == 0 && pullFailed {
					log.Printf("sould cluster completely corrupted")
					httpStatus = http.StatusServiceUnavailable
				}
			}
		}

		writeResponseError(
			response, httpStatus,
			strings.Join(responseMessages, "\n\n"),
		)
	}
}

func (server MirrorServer) HandleGET(
	response http.ResponseWriter,
	request *http.Request,
) {
	mirrorName := strings.Trim(request.RequestURI, "/")

	mirror, err := GetMirror(server.GetStorageDir(), mirrorName)

	if err != nil {
		err = fmt.Errorf(
			"can't get mirror '%s': %s",
			mirrorName, err.Error(),
		)

		log.Println(err)

		httpStatus := http.StatusInternalServerError
		if os.IsNotExist(err) {
			httpStatus = http.StatusNotFound
		}

		writeResponseError(
			response, httpStatus, err.Error(),
		)

		return
	}

	mirrorState := server.stateTable.GetState(mirrorName)
	if mirrorState != MirrorStateSuccess {
		log.Printf(
			"mirror '%s' state is %s, trying to pull changes",
			mirrorName, mirrorState,
		)

		newMirrorState := MirrorStateSuccess

		err = mirror.Pull()
		if err != nil {
			log.Printf(
				"can't pull mirror '%s': %s",
				mirrorName, err.Error(),
			)

			newMirrorState = MirrorStateFailed
		}

		mirrorState = newMirrorState
		server.stateTable.SetState(mirrorName, mirrorState)
	}

	response.Header().Set("X-State", string(mirrorState))

	if mirrorState != MirrorStateSuccess {
		modDate, err := mirror.GetModDate()
		if err != nil {
			err = fmt.Errorf(
				"could not get modify time of '%s' mirror repository: %s",
				mirrorName, err.Error(),
			)
		}

		response.Header().Set("X-Date", modDate.UTC().Format(http.TimeFormat))
	}

	archive, err := mirror.GetArchive()
	if err != nil {
		log.Printf(
			"can't get tar archive of '%s' mirror: %s",
			mirrorName, err.Error(),
		)
		writeResponseError(
			response, http.StatusInternalServerError, err.Error(),
		)

		return
	}

	response.Header().Set("Content-Type", "application/x-tar")

	_, err = response.Write(archive)
	if err != nil {
		log.Printf(
			"got error while writing archive output (mirror: %s): %s",
			mirrorName, archive,
		)

		writeResponseError(
			response, http.StatusInternalServerError, err.Error(),
		)
	}
}

// GetMirror will try to get mirror from server storage director if can not,
// then will try to create mirror with passed arguments.
// if had to create mirror, then 'created bool' returning variable will be true.
func (server MirrorServer) GetMirror(
	name string, origin string,
) (mirror Mirror, created bool, err error) {
	mirror, err = GetMirror(
		server.GetStorageDir(), name,
	)
	if err != nil {
		if !os.IsNotExist(err) {
			return mirror, created, fmt.Errorf(
				"can't get mirror '%s': %s",
				name, err.Error(),
			)
		}

		mirror, err = CreateMirror(
			server.GetStorageDir(), name, origin,
		)
		if err != nil {
			return mirror, created, fmt.Errorf(
				"can't create mirror '%s': %s",
				name, err.Error(),
			)
		}

		return mirror, true, nil
	}

	// if mirror is already exists
	actualOrigin, err := mirror.GetOrigin()
	if err != nil {
		err = fmt.Errorf(
			"can't get origin of mirror '%s' : %s",
			name, err.Error(),
		)
	} else if actualOrigin != origin {
		err = fmt.Errorf(
			"mirror '%s' have another origin, actual = '%s'",
			name, actualOrigin,
		)
	}

	return mirror, false, err
}

func writeResponseError(
	response http.ResponseWriter,
	httpStatusCode int,
	message string,
) {
	response.WriteHeader(httpStatusCode)

	encodedMessage, err := json.Marshal(
		map[string]interface{}{"error": message},
	)
	if err != nil {
		panic(err)
	}

	response.Write(encodedMessage)
}
