package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type TarRequest struct {
	MirrorName string
	Reference  string
}

func (request TarRequest) String() string {
	return fmt.Sprintf(
		"TAR name='%s' reference='%s'",
		request.MirrorName, request.Reference,
	)
}

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
