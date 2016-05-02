package main

import (
	"net/http"
)

type PropagatableRequest interface {
	GetHTTPRequest(MirrorSlave) (*http.Request, error)
}
