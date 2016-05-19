package main

import (
	"net/http"
	"net/url"
)

type StatusRequest struct {
	format string
}

func (request StatusRequest) FormatJSON() bool {
	return request.format == "json"
}

func (request StatusRequest) FormatTOML() bool {
	return request.format == "toml"
}

func (request StatusRequest) FormatHierarchical() bool {
	return request.format == "hierarchical"
}

func (request StatusRequest) GetHTTPRequest(
	slave MirrorSlave,
) (*http.Request, error) {
	return http.NewRequest(
		"GET",
		"http://"+string(slave)+"/x/status?format="+request.format,
		nil,
	)
}

func ExtractStatusRequest(url *url.URL) StatusRequest {
	var format string

	formatValue := url.Query().Get("format")
	switch formatValue {
	case "json", "toml":
		format = formatValue

	default:
		format = "hierarchical"
	}

	return StatusRequest{
		format: format,
	}
}
