package main

import "net/http"

// StatusRequest is request which should be handled by all sould servers
// independently of server role and servers should response their mirrors
// statuses, if server is master then server should report about all slave
// servers and their statuses.
type StatusRequest struct {
	format string
}

// IsFormatJSON returns true if server should response in JSON format.
func (request StatusRequest) IsFormatJSON() bool {
	return request.format == "json"
}

// IsFormatTOML returns true if server should response in TOML format.
func (request StatusRequest) IsFormatTOML() bool {
	return request.format == "toml"
}

// IsFormatHierarchical returns true if server should response in hierarchical
// format. Really human-friendly format.
func (request StatusRequest) IsFormatHierarchical() bool {
	return request.format == "hierarchical"
}

// GetHTTPRequest returns http request which can be propagated to other
// servers.
func (request StatusRequest) GetHTTPRequest(
	slave SecondaryServer,
) (*http.Request, error) {
	return http.NewRequest(
		"GET",
		"http://"+string(slave)+"/x/status?format=json",
		nil,
	)
}
