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

func oneLineError(err error) string {
	str := err.Error()
	lines := strings.SplitN(str, "\n", 2)
	return lines[0]
}
