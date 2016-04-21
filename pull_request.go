package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/ajg/form"
)

type PullRequest struct {
	MirrorName     string `form:"name"`
	MirrorOrigin   string `form:"origin"`
	Spoof          bool   `form:"spoof,omitempty"`
	SpoofingBranch string `form:"branch,omitempty"`
	SpoofingTag    string `form:"tag,omitempty"`
}

func (request PullRequest) String() string {
	return fmt.Sprintf(
		"PULL name = '%s' origin = '%s' spoof = %t branch = '%s' tag = '%s'",
		request.MirrorName, request.MirrorOrigin,
		request.Spoof, request.SpoofingBranch, request.SpoofingTag,
	)
}

func ExtractPullRequest(
	values url.Values, insecure bool,
) (PullRequest, error) {
	var request PullRequest
	err := form.DecodeValues(&request, values)
	if err != nil {
		return request, err
	}

	if request.Spoof {
		values.Set("branch", values.Get("branch"))
		values.Set("tag", values.Get("tag"))
	}

	err = validateValues(values)
	if err != nil {
		return request, err
	}

	if !insecure && !isURL(request.MirrorOrigin) {
		return request, errors.New("field 'origin' has an invalid URL")
	}

	return request, err
}

func validateValues(values url.Values) error {
	for key, _ := range values {
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
