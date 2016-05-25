package main

import (
	"net/http"
	"net/url"
)

// StatusRequest is request which should be handled by all sould servers
// independently of server role and servers should response their mirrors
// statuses, if server is master then server should report about all slave
// servers and their statuses.
type StatusRequest struct {
	format string
}

// FormatJSON returns true if server should response in JSON format.
func (request StatusRequest) FormatJSON() bool {
	return request.format == "toml"
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
	slave MirrorSlave,
) (*http.Request, error) {
	return http.NewRequest(
		"GET",
		"http://"+string(slave)+"/x/status?format=json",
		nil,
	)
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
