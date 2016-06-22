package main

import "net/http"

// StatusRequest is request which should be handled by all sould servers
// independently of server role and servers should response their mirrors
// statuses, if server is master then server should report about all slave
// servers and their statuses.
type StatusRequest struct {
	format string
}

// FormatJSON returns true if server should response in JSON format.
func (request StatusRequest) FormatJSON() bool {
	return request.format == "json"
}

// FormatTOML returns true if server should response in TOML format.
func (request StatusRequest) FormatTOML() bool {
	return request.format == "toml"
}

// FormatHierarchical returns true if server should response in hierarchical
// format. Really human-friendly format.
func (request StatusRequest) FormatHierarchical() bool {
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
