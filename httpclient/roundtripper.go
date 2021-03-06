package httpclient

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrStatusCode is a sentinel error for detecting if the request has
	// been rejected because of a non-acceptable status code
	ErrStatusCode = errors.New("the status code returned is not acceptable for the request method")

	// ErrServerError is the error that is returned when the server is returning a server error
	ErrServerError = errors.New("the server is returning a server error")
)

// AllowableStatusCodes is a map of the allowable status codes for a specific request method
var AllowableStatusCodes = map[string]StatusSet{
	http.MethodGet:     {http.StatusOK: {}, http.StatusAccepted: {}, http.StatusNoContent: {}},
	http.MethodHead:    {http.StatusOK: {}, http.StatusNoContent: {}},
	http.MethodPost:    {http.StatusOK: {}, http.StatusCreated: {}, http.StatusAccepted: {}, http.StatusNoContent: {}},
	http.MethodPut:     {http.StatusOK: {}, http.StatusAccepted: {}, http.StatusCreated: {}, http.StatusNoContent: {}},
	http.MethodPatch:   {http.StatusOK: {}, http.StatusAccepted: {}, http.StatusCreated: {}, http.StatusNoContent: {}},
	http.MethodDelete:  {http.StatusOK: {}, http.StatusAccepted: {}, http.StatusNoContent: {}},
	http.MethodConnect: {http.StatusOK: {}, http.StatusAccepted: {}, http.StatusNoContent: {}},
	http.MethodTrace:   {http.StatusOK: {}, http.StatusAccepted: {}, http.StatusNoContent: {}},
}

// StatusCode is an http StatusCode
type StatusCode int

// String implements fmt.Stringer
func (s StatusCode) String() string {
	return http.StatusText(int(s))
}

// StatusSet is a set for checking for allowable status codes
type StatusSet map[StatusCode]struct{}

// NewStatusCheckingTripper creates a new status checking tripper to wrap a roundtripper
func NewStatusCheckingTripper(tripper http.RoundTripper) StatusCheckingTripper {
	return StatusCheckingTripper{
		AllowableStatus: AllowableStatusCodes,
		Base:            tripper,
	}
}

// StatusCheckingTripper is a round tripper that checks for
type StatusCheckingTripper struct {
	AllowableStatus map[string]StatusSet
	Base            http.RoundTripper
}

// RoundTrip implements http.RoundTripper
func (s StatusCheckingTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	allowableCodes, ok := s.AllowableStatus[request.Method]
	if !ok {
		return s.Base.RoundTrip(request)
	}
	response, err := s.Base.RoundTrip(request)
	if err != nil {
		return response, fmt.Errorf("could not process request to check status, got (%w)", err)
	}
	status := StatusCode(response.StatusCode)
	if _, ok := allowableCodes[status]; !ok {
		return response, fmt.Errorf(
			"(%w) (%s), is not an acceptable status for method (%s)",
			handleBadStatusCode(response.StatusCode),
			status,
			request.Method,
		)
	}
	return response, nil
}

func handleBadStatusCode(status int) error {
	switch {
	case status >= http.StatusInternalServerError:
		return ErrServerError
	default:
		return ErrStatusCode
	}
}
