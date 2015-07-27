package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func (server *MirrorServer) ListenHTTP() error {
	http.Handle("/", server)

	return http.ListenAndServe(server.GetListenAddress(), nil)
}

func (server *MirrorServer) ServeHTTP(
	response http.ResponseWriter, request *http.Request,
) {
	defer func() {
		panicError := recover()
		if panicError != nil {
			log.Println(panicError)
		}
	}()

	method := strings.ToUpper(request.Method)

	switch method {
	case "POST":
		server.handlePOST(response, request)

	case "GET":
		server.handleGET(response, request)

	default:
		response.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("got request with unsupported method: '%s'", method)
	}
}

func (server *MirrorServer) handlePOST(
	response http.ResponseWriter, request *http.Request,
) {
	var (
		responseMessages []string

		hadToCreateMirror   bool
		pullFailed          bool
		forwardingFailed    bool
		forwardingFailedAll bool
	)

	mirrorName, mirrorOrigin, err := getMirrorParams(request)
	if err != nil {
		log.Println("%s, got http form: %#v", err, request.Form)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf(
		"got pull request for mirror, name = '%s', origin = '%s'",
		mirrorName, mirrorOrigin,
	)

	if !server.insecureMode && !isURL(mirrorOrigin) {
		message := "mirror origin should be URL"
		log.Printf("%s, url is '%s'", message, mirrorOrigin)
		http.Error(response, message, http.StatusForbidden)
		return
	}

	mirror, hadToCreateMirror, err := server.GetMirror(mirrorName, mirrorOrigin)
	if err != nil {
		log.Println(err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if hadToCreateMirror {
		log.Printf(
			"mirror '%s' successfully created",
			mirrorName,
		)
	}

	err = mirror.Pull()
	if err != nil {
		pullFailed = true

		message := fmt.Sprintf(
			"can't update mirror '%s': %s", mirrorName, err.Error(),
		)

		log.Println(message)
		responseMessages = append(responseMessages, message)

		server.stateTable.SetState(mirrorName, MirrorStateFailed)
	} else {
		log.Printf("mirror '%s' successfully updated", mirrorName)

		server.stateTable.SetState(mirrorName, MirrorStateSuccess)
	}

	// if an error occurred during mirror pull, master sould should be
	// tolerant and try to forward request to slaves

	if server.IsMaster() {
		slaves := server.GetMirrorUpstream()
		if len(slaves) > 0 {
			updatedSlaves, errors := slaves.Pull(
				mirrorName, mirrorOrigin, server.httpClient,
			)

			if len(updatedSlaves) > 0 {
				log.Printf(
					"request successfully propagated to slaves %s",
					strings.Join(updatedSlaves.GetHosts(), ", "),
				)
			}

			if len(errors) > 0 {
				forwardingFailed = true
				if len(updatedSlaves) == 0 {
					forwardingFailedAll = true
				}

				for _, err := range errors {
					log.Println(err)
					responseMessages = append(responseMessages, err.Error())
				}
			}
		}
	}

	var status int
	switch {
	case forwardingFailedAll && pullFailed:
		status = http.StatusServiceUnavailable
		log.Printf("sould cluster completely corrupted")

	case forwardingFailed:
		status = http.StatusBadGateway

	case pullFailed:
		status = http.StatusInternalServerError

	case hadToCreateMirror:
		status = http.StatusCreated

	default:
		status = http.StatusOK
	}

	http.Error(response, strings.Join(responseMessages, "\n\n"), status)
}

func (server MirrorServer) handleGET(
	response http.ResponseWriter, request *http.Request,
) {
	mirrorName := strings.Trim(request.URL.Path, "/")

	reference := request.URL.Query().Get("ref")
	if reference == "" {
		reference = "master"
	}

	log.Printf(
		"got get tar archive request for mirror '%s', ref '%s'",
		mirrorName, reference,
	)

	mirror, err := GetMirror(server.GetStorageDir(), mirrorName)
	if err != nil {
		var status int
		if os.IsNotExist(err) {
			status = http.StatusNotFound
		} else {
			status = http.StatusInternalServerError
		}

		message := fmt.Sprintf(
			"can't get mirror '%s': %s", mirrorName, err.Error(),
		)

		log.Println(message)
		http.Error(response, message, status)

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
			newMirrorState = MirrorStateFailed

			log.Printf(
				"can't pull mirror '%s': %s",
				mirrorName, err.Error(),
			)
		}

		server.stateTable.SetState(mirrorName, newMirrorState)
		mirrorState = newMirrorState
	}

	modifyDate, err := mirror.GetModifyDate()
	if err != nil {
		message := fmt.Sprintf(
			"could not get modify time of '%s' mirror repository: %s",
			mirrorName, err.Error(),
		)

		log.Printf(message)
		http.Error(response, message, http.StatusInternalServerError)
		return
	}

	response.Header().Set("X-State", mirrorState.String())
	response.Header().Set("X-Date", modifyDate.UTC().Format(http.TimeFormat))

	archive, err := mirror.GetArchive(reference)
	if err != nil {
		log.Printf(
			"can't get tar archive of '%s' mirror: %s",
			mirrorName, err.Error(),
		)

		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/x-tar")

	_, err = response.Write(archive)
	if err != nil {
		log.Printf(
			"got error while writing archive output (mirror: %s): %s",
			mirrorName, archive,
		)

		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetMirror will try to get mirror from server storage directory and if can
// not, then will try to create mirror with passed arguments.
func (server MirrorServer) GetMirror(
	name string, origin string,
) (mirror Mirror, hasBeenCreated bool, err error) {
	mirror, err = GetMirror(server.GetStorageDir(), name)
	if err != nil {
		if !os.IsNotExist(err) {
			return Mirror{}, false, fmt.Errorf(
				"can't get mirror '%s': %s", name, err.Error(),
			)
		}

		mirror, err = CreateMirror(server.GetStorageDir(), name, origin)
		if err != nil {
			// hasBeenCreated variable should be false, because creating is failed.
			return Mirror{}, false, fmt.Errorf(
				"can't create mirror '%s': %s", name, err.Error(),
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

func getMirrorParams(
	request *http.Request,
) (name string, origin string, err error) {
	err = request.ParseForm()
	if err != nil {
		return "", "", err
	}

	name = request.FormValue("name")
	origin = request.FormValue("origin")

	switch {
	case name == "":
		err = fmt.Errorf("mirror name can not be empty")

	case origin == "":
		err = fmt.Errorf("mirror origin (clone url) can not be empty")
	}

	return name, origin, err
}

func isURL(str string) bool {
	_, err := url.Parse(str)
	if err != nil {
		return false
	}

	var prefixes = []string{
		"ssh://", "https://", "http://", "git://",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}

	return false
}
