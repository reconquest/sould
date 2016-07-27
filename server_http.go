package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/ajg/form"
)

// ListenHTTP starts a new http (tcp) listener at specified listening address.
func (server *Server) ListenHTTP() error {
	http.Handle("/", server)

	return http.ListenAndServe(server.GetListenAddress(), nil)
}

// ServeHTTP is entrypoint of all HTTP connections with sould server.
func (server *Server) ServeHTTP(
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

// ExtractPullRequest parses post form and creates new instance of PullRequest,
// if insecure is false (by default) then ExtractPullRequest will check that
// given mirror origin url is really url.
func ExtractPullRequest(
	values url.Values, insecure bool,
) (*PullRequest, error) {
	var request PullRequest
	err := form.DecodeValues(&request, values)
	if err != nil {
		return nil, err
	}

	if request.Spoof {
		values.Set("branch", values.Get("branch"))
		values.Set("tag", values.Get("tag"))
	}

	err = validateValues(values)
	if err != nil {
		return nil, err
	}

	if !insecure && !isURL(request.MirrorOrigin) {
		return nil, errors.New("field 'origin' has an invalid URL")
	}

	return &request, err
}

func validateValues(values url.Values) error {
	for key := range values {
		value := strings.TrimSpace(values.Get(key))
		if value == "" {
			return errors.New("field '" + key + "' has an empty value")
		}

		if strings.HasPrefix(value, "-") {
			return errors.New("field '" + key + "' has an insecure value")
		}
	}

	return nil
}

// ExtractTarRequest parses given URL and extracts request for downloading tar
// archive.
func ExtractTarRequest(url *url.URL) (TarRequest, error) {
	request := TarRequest{
		MirrorName: strings.Trim(url.Path, "/"),
		Reference:  strings.TrimSpace(url.Query().Get("ref")),
	}

	if request.MirrorName == "" {
		return request, errors.New("field 'name' has an empty value")
	}

	if request.Reference == "" {
		request.Reference = "master"
	}

	return request, nil
}

// ExtractStatusRequest returns instance of StatusRequest basing on specified
// URL.
func ExtractStatusRequest(url *url.URL) StatusRequest {
	var format string

	formatValue := url.Query().Get("format")
	switch formatValue {
	case "toml", "json":
		format = formatValue

	default:
		format = "hierarchical"
	}

	return StatusRequest{
		format: format,
	}
}
