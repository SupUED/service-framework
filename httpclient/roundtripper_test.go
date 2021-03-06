package httpclient_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BlackBX/service-framework/httpclient"
)

type roundTripperFunc func(r *http.Request) (*http.Response, error)

func (r roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return r(request)
}

func TestStatusCheckingTripper_RoundTripSucceeds(t *testing.T) {
	roundTripper := httpclient.NewStatusCheckingTripper(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		resp := &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Proto:      "https",
			Body:       http.NoBody,
		}
		return resp, nil
	}))
	req := httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
	resp, err := roundTripper.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
}

func TestStatusCheckingTripper_RoundTripFails(t *testing.T) {
	roundTripper := httpclient.NewStatusCheckingTripper(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     http.StatusText(http.StatusBadRequest),
			StatusCode: http.StatusBadRequest,
			Proto:      "https",
		}, nil
	}))
	req, err := http.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	client := http.DefaultClient
	client.Transport = roundTripper
	resp, err := client.Do(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected an error got none")
	}
	if !errors.Is(err, httpclient.ErrStatusCode) {
		t.Fatalf("expected error to be a status code error (%s)", err)
	}
}

func TestStatusCheckingTripper_RoundTripFailsServerError(t *testing.T) {
	roundTripper := httpclient.NewStatusCheckingTripper(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     http.StatusText(http.StatusInternalServerError),
			StatusCode: http.StatusInternalServerError,
			Proto:      "https",
		}, nil
	}))
	req, err := http.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	client := http.DefaultClient
	client.Transport = roundTripper
	resp, err := client.Do(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected an error got none")
	}
	if !errors.Is(err, httpclient.ErrServerError) {
		t.Fatalf("expected error to be a status code error (%s)", err)
	}
}

func TestStatusCheckingTripper_CustomMethodSucceeds(t *testing.T) {
	roundTripper := httpclient.NewStatusCheckingTripper(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     http.StatusText(http.StatusBadRequest),
			StatusCode: http.StatusBadRequest,
			Proto:      "https",
			Body:       http.NoBody,
		}, nil
	}))
	req := httptest.NewRequest("CUSTOM", "https://example.com", http.NoBody)
	resp, err := roundTripper.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
}

func TestStatusCheckingTripper_RoundTripFailsWhenTripperFails(t *testing.T) {
	roundTripper := httpclient.NewStatusCheckingTripper(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Proto:      "https",
			Body:       http.NoBody,
		}, errors.New("an error")
	}))
	req := httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
	resp, err := roundTripper.RoundTrip(req)
	if err == nil {
		t.Fatal("expected an error got none")
	}
	_ = resp.Body.Close()
}
