package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
		mirrorName     = request.FormValue("name")
		mirrorCloneURL = request.FormValue("url")
	)

	for paramName, paramValue := range map[string]string{
		"name": mirrorName,
		"url":  mirrorCloneURL,
	} {
		if paramValue == "" {
			emptyErr := fmt.Errorf(
				"'%s' param not found in request", paramName,
			)

			log.Printf("%s, request = '%#v'", emptyErr, request.Form)

			writeResponseError(
				response, http.StatusBadRequest,
				emptyErr.Error(),
			)

			return
		}
	}

	log.Println(
		"got pull request for mirror with name = %s, clone url = %s",
		mirrorName, mirrorCloneURL,
	)

	mirrorFound := true

	mirror, err := GetMirror(
		server.GetStorageDir(), mirrorName, mirrorCloneURL,
	)

	if err != nil {
		if err == ErrMirrorNotFound {
			mirrorFound = false

			log.Printf(
				"mirror '%s' not found, trying to create",
				mirrorName,
			)

			var createErr error
			mirror, createErr = CreateMirror(
				server.GetStorageDir(), mirrorName, mirrorCloneURL,
			)
			if createErr != nil {
				err = fmt.Errorf(
					"can't create mirror '%s': %s",
					mirrorName, createErr.Error(),
				)

				log.Println(err)
				writeResponseError(
					response, http.StatusInternalServerError,
					err.Error(),
				)

				return
			}

			log.Println(
				"mirror '%s' successfully created",
				mirrorName,
			)
		} else {
			err = fmt.Errorf(
				"can't get mirror '%s': %s",
				mirrorName, err.Error(),
			)

			log.Println(err)
			writeResponseError(
				response, http.StatusInternalServerError, err.Error(),
			)
		}
	}

	var (
		responseMessages []string

		httpStatus = http.StatusOK
	)

	pullErr := mirror.Pull()
	if pullErr != nil {
		server.pullStateTable.SetState(mirrorName, PullStateFailed)

		httpStatus = http.StatusInternalServerError

		pullErr = fmt.Errorf(
			"can't update mirror '%s': %s",
			mirrorName, pullErr.Error(),
		)

		log.Println(pullErr)

		// an error occurred during mirror pull, master sould should be
		// tolerant and try to forward request to slaves
		if server.IsMaster() && mirrorFound {
			responseMessages = append(responseMessages, pullErr.Error())
		} else {
			writeResponseError(response, httpStatus, pullErr.Error())

			return
		}
	} else {
		server.pullStateTable.SetState(mirrorName, PullStateSuccess)

		log.Println("mirror '%s' successfully updated", mirrorName)
	}

	if server.IsMaster() {
		slaves := server.GetSlaves()
		if len(slaves) > 0 {
			fedSlaves, errors := feedSlaves(
				slaves, server.httpClient,
				mirrorName, mirrorCloneURL,
			)

			if len(fedSlaves) > 0 {
				log.Println(
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

				if len(fedSlaves) == 0 && pullErr != nil {
					log.Println("sould cluster completely corrupted")
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
	//var (
	//response []byte
	//err      error
	//)

	//response, err = mirror.GetTarArchive()

	//if err != nil {
	//writeResponseError(
	//response,
	//http.StatusInternalServerError,
	//err.Error(),
	//)
	//return
	//}

	//response.Header().Set("Content-Type", "application/x-tar")

	//_, err = response.Write(response)
	//if err != nil {
	//writeResponseError(
	//response,
	//http.StatusInternalServerError,
	//err.Error(),
	//)
	//}
}

func writeResponseError(
	response http.ResponseWriter,
	httpStatusCode int,
	message string,
) {
	response.WriteHeader(httpStatusCode)

	log.Println(message)

	encodedMessage, err := json.Marshal(
		map[string]interface{}{"error": message},
	)
	if err != nil {
		panic(err)
	}

	response.Write(encodedMessage)
}
