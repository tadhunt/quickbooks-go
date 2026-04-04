// Copyright (c) 2020, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

var (
	// ErrNotFound is returned when a query returns no results.
	ErrNotFound = errors.New("not found")

	// ErrMissingID is returned when a required ID or sync token is not provided.
	ErrMissingID = errors.New("missing id")

	// ErrNoDownloadURL is returned when an attachable has no download URL.
	ErrNoDownloadURL = errors.New("no download URL returned")

	// ErrRateLimit is returned when the API rate limit is exceeded.
	ErrRateLimit = errors.New("waiting for rate limit")
)

// HTTPError represents an unexpected HTTP status code.
// Use errors.As to extract the status code from a wrapped error.
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: status %d", e.Message, e.StatusCode)
}

// Failure is the outermost struct that holds an error response.
type Failure struct {
	Fault struct {
		Error []struct {
			Message string
			Detail  string
			Code    string `json:"code"`
			Element string `json:"element"`
		}
		Type string `json:"type"`
	}
	Time Date `json:"time"`
}

// Error implements the error interface.
func (f Failure) Error() string {
	text, err := json.Marshal(f)
	if err != nil {
		return fmt.Sprintf("unexpected error while marshalling error: %v", err)
	}

	return string(text)
}

// parseFailure takes a response reader and tries to parse a Failure.
func parseFailure(resp *http.Response) error {
	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("When reading response body:" + err.Error())
	}

	var errStruct Failure

	if err = json.Unmarshal(msg, &errStruct); err != nil {
		return errors.New(strconv.Itoa(resp.StatusCode) + " " + string(msg))
	}

	return errStruct
}
