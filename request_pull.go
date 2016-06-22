package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/ajg/form"
)

// PullRequest is the request for pulling changeset from remote repository
type PullRequest struct {
	// MirrorName is name which will be used for identify repository mirror.
	MirrorName string `form:"name"`

	// MirrorOrigin is clone/fetch URL of remote repository.
	MirrorOrigin string `form:"origin"`

	// Spoof is positional parameter which needed for pre-receive feature with
	// spoofing changesets.
	Spoof bool `form:"spoof,omitempty"`

	// SpoofingBranch identifies branch which will be spoofed.
	SpoofingBranch string `form:"branch,omitempty"`

	// SpoofingTag identifies tag which will be spoofed.
	SpoofingTag string `form:"tag,omitempty"`
}

func (request *PullRequest) String() string {
	return fmt.Sprintf(
		"PULL name = '%s' origin = '%s' spoof = %t branch = '%s' tag = '%s'",
		request.MirrorName, request.MirrorOrigin,
		request.Spoof, request.SpoofingBranch, request.SpoofingTag,
	)
}

// GetHTTPRequest which can be executed basing on given pull request data.
func (request *PullRequest) GetHTTPRequest(
	slave ServerFollowerServer,
) (*http.Request, error) {
	payload, err := form.EncodeToString(request)
	if err != nil {
		return nil, NewError(err, "can't create payload")
	}

	httpRequest, err := http.NewRequest(
		"POST", "http://"+string(slave)+"/", bytes.NewBufferString(payload),
	)
	if err != nil {
		return nil, err
	}

	httpRequest.Header.Set(
		"Content-Type", "application/x-www-form-urlencoded",
	)

	return httpRequest, nil
}
