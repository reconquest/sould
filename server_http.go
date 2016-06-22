package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// ListenHTTP starts a new http (tcp) listener at specified listening address.
func (server *ServerHTTP) ListenHTTP() error {
	http.Handle("/", server)

	return http.ListenAndServe(server.GetListenAddress(), nil)
}

// ServeHTTP is entrypoint of all HTTP connections with sould server.
func (server *ServerHTTP) ServeHTTP(
	response http.ResponseWriter, request *http.Request,
) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Errorf("(http) PANIC: %s\n%#v\n%s", err, request, stack())
		}
	}()

	logRequest(request)

	var method = request.Method
	switch {
	case method == "POST":
		pullRequest, err := ExtractPullRequest(
			request.Form, server.insecureMode,
		)
		if err != nil {
			logger.Error(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info(pullRequest)

		server.HandlePullRequest(response, pullRequest)

	case method == "GET" &&
		strings.TrimRight(request.URL.Path, "/") == "/x/status":

		statusRequest := ExtractStatusRequest(request.URL)

		logger.Info(statusRequest)

		server.HandleStatusRequest(response, statusRequest)

	case method == "GET":
		tarRequest, err := ExtractTarRequest(request.URL)
		if err != nil {
			logger.Error(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info(tarRequest)

		server.HandleTarRequest(response, tarRequest)

	default:
		response.WriteHeader(http.StatusMethodNotAllowed)
		logger.Errorf("unsupported method: %s", request.Method)
	}
}

// GetMirror returns existing mirror or creates new instance in storage
// directory.
func (server *ServerHTTP) GetMirror(
	name string, origin string,
) (mirror Mirror, created bool, err error) {
	mirror, err = GetMirror(server.GetStorageDir(), name)
	if err != nil {
		if !os.IsNotExist(err) {
			return Mirror{}, false, err
		}

		mirror, err = CreateMirror(server.GetStorageDir(), name, origin)
		if err != nil {
			return Mirror{}, false, NewError(err, "can't create new mirror")
		}

		return mirror, true, nil
	}

	mirrorURL, err := mirror.GetURL()
	if err != nil {
		return mirror, false, NewError(err, "can't get mirror origin url")
	}

	if mirrorURL != origin {
		return mirror, false, fmt.Errorf(
			"mirror have different origin url (%s)",
			mirrorURL,
		)
	}

	return mirror, true, nil
}
