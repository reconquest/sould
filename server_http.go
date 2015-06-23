package main

import (
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
		writeResponseMessage(response, http.StatusBadRequest, err.Error())

		return
	}

	var (
		responseMessages []string

		hadCreateMirror     bool
		pullFailed          bool
		forwardingFailed    bool
		forwardingFailedAll bool

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

			writeResponseMessage(
				response, http.StatusBadRequest, err.Error(),
			)

			return
		}
	}

	log.Printf(
		"got pull request for mirror with name = %s, origin = %s",
		mirrorName, mirrorOrigin,
	)

	if !server.unsecure && !isURL(mirrorOrigin) {
		err = fmt.Errorf("mirror origin should be URL")

		log.Println(err)
		writeResponseMessage(
			response, http.StatusForbidden, err.Error(),
		)

		return
	}

	mirror, hadCreateMirror, err := server.GetMirror(mirrorName, mirrorOrigin)
	if err != nil {
		log.Println(err)
		writeResponseMessage(
			response, http.StatusInternalServerError, err.Error(),
		)

		return
	}

	if hadCreateMirror {
		log.Printf(
			"mirror '%s' successfully created",
			mirrorName,
		)
	}

	err = mirror.Pull()
	if err != nil {
		pullFailed = true

		err = fmt.Errorf(
			"can't update mirror '%s': %s", mirrorName, err.Error(),
		)
		log.Println(err)
		responseMessages = append(responseMessages, err.Error())

		server.stateTable.SetState(mirrorName, MirrorStateFailed)
	} else {
		log.Printf("mirror '%s' successfully updated", mirrorName)

		server.stateTable.SetState(mirrorName, MirrorStateSuccess)
	}

	// if an error occurred during mirror pull, master sould should be
	// tolerant and try to forward request to slaves

	if server.IsMaster() {
		slaves := server.GetSlaves()
		log.Printf("slaves: %#v", slaves)
		if len(slaves) > 0 {
			fedSlaves, errors := feedSlaves(
				slaves, server.httpClient,
				mirrorName, mirrorOrigin,
			)
			log.Printf("fedSlaves: %#v", fedSlaves)
			log.Printf("errors: %#v", errors)

			if len(fedSlaves) > 0 {
				log.Printf(
					"request successfully forwarded to slaves %s",
					strings.Join(fedSlaves, ", "),
				)
			}

			if len(errors) > 0 {
				forwardingFailed = true

				for _, err := range errors {
					log.Println(err)

					responseMessages = append(responseMessages, err.Error())
				}

				if len(fedSlaves) == 0 {
					forwardingFailedAll = true
				}
			}
		}
	}

	var httpStatus int
	switch {
	case forwardingFailedAll && pullFailed:
		log.Printf("sould cluster completely corrupted")
		httpStatus = http.StatusServiceUnavailable

	case forwardingFailed:
		httpStatus = http.StatusBadGateway

	case pullFailed:
		httpStatus = http.StatusInternalServerError

	case hadCreateMirror:
		httpStatus = http.StatusCreated

	default:
		httpStatus = http.StatusOK
	}

	writeResponseMessage(
		response, httpStatus, strings.Join(responseMessages, "\n\n"),
	)
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

		writeResponseMessage(
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
		writeResponseMessage(
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

		writeResponseMessage(
			response, http.StatusInternalServerError, err.Error(),
		)
	}
}

// GetMirror will try to get mirror from server storage directory and if can not,
// then will try to create mirror with passed arguments.
func (server MirrorServer) GetMirror(
	name string, origin string,
) (mirror Mirror, hadCreate bool, err error) {
	mirror, err = GetMirror(
		server.GetStorageDir(), name,
	)
	if err != nil {
		if !os.IsNotExist(err) {
			return mirror, false, fmt.Errorf(
				"can't get mirror '%s': %s",
				name, err.Error(),
			)
		}

		mirror, err = CreateMirror(
			server.GetStorageDir(), name, origin,
		)
		if err != nil {
			return mirror, true, fmt.Errorf(
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

func writeResponseMessage(
	response http.ResponseWriter, httpStatus int, message string,
) {
	response.WriteHeader(httpStatus)
	response.Write([]byte(message))
}

func isURL(origin string) bool {
	prefixes := []string{
		"ssh://",
		"https://",
		"http://",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(origin, prefix) {
			return true
		}
	}

	return false
}
