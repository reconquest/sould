package main

import (
	"net/http"
	"net/url"
	"strings"
)

func isURL(str string) bool {
	_, err := url.Parse(str)
	if err != nil {
		return false
	}

	var prefixes = []string{
		"ssh://", "https://", "http://", "git://",
		"ssh+git://", "git+ssh://",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}

	return false
}

func logRequest(request *http.Request) {
	var (
		parsingErr = request.ParseForm()

		format = "%s %s '%s' %s"
		values = []interface{}{
			request.Method,
			request.URL,
			request.Form.Encode(),
			request.RemoteAddr,
		}
	)

	if parsingErr != nil {
		format += ": %s"
		values = append(values, parsingErr)

		logger.Errorf(format, values...)
	} else {
		logger.Infof(format, values...)
	}

}

func logPropagation(request string, propagation *RequestPropagation) {
	var (
		successes = propagation.ResponsesSuccess()
		errors    = propagation.ResponsesError()
		total     = len(successes) + len(errors)
	)

	logger.Infof(
		"%s request propagated to %d slaves, "+
			"success %v (%.2f%%), error %v (%.2f%%)",
		request, total,
		len(successes), percent(len(successes), total),
		len(errors), percent(len(errors), total),
	)

	if len(successes) > 0 {
		logger.Infof(
			"%s request successfully propagated to %s",
			request, strings.Join(successes.GetHosts(), ", "),
		)
	}

	for _, err := range errors {
		logger.Errorf(
			"slave %s propagating %s request error: %s",
			err.Slave, request, err.Error(),
		)
	}
}

func oneLineError(err error) string {
	str := err.Error()
	lines := strings.SplitN(str, "\n", 2)
	return lines[0]
}

func percent(complete int, total int) float64 {
	if total == 0 {
		return 0
	}

	return float64(complete*100) / float64(total)
}
