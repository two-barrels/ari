// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRequestStopsAtCustomHTTPClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 25 * time.Millisecond}
	client := New(&Options{
		URL:        server.URL,
		HTTPClient: httpClient,
	})

	startedAt := time.Now()
	err := client.get("/channels", nil)

	require.Error(t, err)
	require.Less(t, time.Since(startedAt), 500*time.Millisecond)

	var networkError net.Error
	require.ErrorAs(t, err, &networkError)
	require.True(t, networkError.Timeout())
}

func TestNewUsesProvidedHTTPClient(t *testing.T) {
	httpClient := &http.Client{}

	client := New(&Options{HTTPClient: httpClient})

	require.Same(t, httpClient, client.httpClient)
}

func TestNewUsesPackageDefaultRequestTimeout(t *testing.T) {
	client := New(&Options{})

	require.Equal(t, RequestTimeout, client.httpClient.Timeout)
}

type responseTransport func(*http.Request) (*http.Response, error)

func (transport responseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func TestRequestPreservesHTTPStatusWithEmptyErrorBody(t *testing.T) {
	client := New(&Options{
		URL: "http://asterisk.test/ari",
		HTTPClient: &http.Client{Transport: responseTransport(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Status:     "404 Not Found",
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    request,
			}, nil
		})},
	})
	var result struct{}
	err := client.get("/channels/missing", &result)
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, CodeFromError(err))
}
