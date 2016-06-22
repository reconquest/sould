package main

import (
	"net/http"
)

// PropagatableRequest is such a request which can be propagated to other sould
// servers.
type PropagatableRequest interface {
	// GetHTTPRequest for specified slave server, returns HTTP request which
	// can be executed by methods like http.Do.
	GetHTTPRequest(SecondaryServer) (*http.Request, error)
}
