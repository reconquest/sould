package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

func (request *PullRequest) GetHTTPRequest(
	slave MirrorSlave,
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

// ExtractPullRequest parses post form and creates new instance of PullRequest,
// if insecure is false (by default) then ExtractPullRequest will check that
// given mirror origin url is really url.
func ExtractPullRequest(
	values url.Values, insecure bool,
) (*PullRequest, error) {
	var request PullRequest
	err := form.DecodeValues(&request, values)
	if err != nil {
		return nil, err
	}

	if request.Spoof {
		values.Set("branch", values.Get("branch"))
		values.Set("tag", values.Get("tag"))
	}

	err = validateValues(values)
	if err != nil {
		return nil, err
	}

	if !insecure && !isURL(request.MirrorOrigin) {
		return nil, errors.New("field 'origin' has an invalid URL")
	}

	return &request, err
}

func validateValues(values url.Values) error {
	for key := range values {
		value := strings.TrimSpace(values.Get(key))
		if value == "" {
			return errors.New("field '" + key + "' has an empty value")
		}

		if strings.HasPrefix(value, "-") {
			return errors.New("field '" + key + "' has an insecure value")
		}
	}

	return nil
}
