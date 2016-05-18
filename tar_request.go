package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// TarRequest is the request for downloading tar archive of specified git
// revision.
type TarRequest struct {
	// MirrorName is a string representation of archiving mirror.
	MirrorName string
	// Reference is a string representation of git revision.
	Reference string
}

func (request TarRequest) String() string {
	return fmt.Sprintf(
		"TAR name='%s' reference='%s'",
		request.MirrorName, request.Reference,
	)
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
