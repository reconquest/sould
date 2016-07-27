package main

import "fmt"

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
