package native

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/two-barrels/ari/v6"
)

type ari23Transport func(*http.Request) (*http.Response, error)

func (transport ari23Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func ari23Response(request *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Header: make(http.Header),
		Body: io.NopCloser(strings.NewReader(body)), Request: request}
}

func TestApplicationFilterEventsWire(t *testing.T) {
	empty := []ari.EventTypeFilter{}
	allowed := []ari.EventTypeFilter{{Type: "StasisStart"}, {Type: "StasisEnd"}}
	disallowed := []ari.EventTypeFilter{{Type: "ChannelDestroyed"}}
	tests := []struct {
		name   string
		filter *ari.ApplicationEventFilter
		want   string
	}{
		{"clear both", nil, `{}`},
		{"clear allowed only", &ari.ApplicationEventFilter{Allowed: &empty}, `{"allowed":[]}`},
		{"set allowed only", &ari.ApplicationEventFilter{Allowed: &allowed}, `{"allowed":[{"type":"StasisStart"},{"type":"StasisEnd"}]}`},
		{"set disallowed only", &ari.ApplicationEventFilter{Disallowed: &disallowed}, `{"disallowed":[{"type":"ChannelDestroyed"}]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(request *http.Request) (*http.Response, error) {
				if request.Method != http.MethodPut || request.URL.EscapedPath() != "/ari/applications/demo/eventFilter" {
					t.Errorf("request = %s %s", request.Method, request.URL.EscapedPath())
				}
				body, err := io.ReadAll(request.Body)
				if err != nil {
					t.Fatal(err)
				}
				var got, want any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte(test.want), &want); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("body = %s, want %s", body, test.want)
				}
				return ari23Response(request, http.StatusOK, `{"name":"demo","events_allowed":[{"type":"StasisStart"}]}`), nil
			})}})
			data, err := client.Application().FilterEvents(ari.NewKey(ari.ApplicationKey, "demo"), test.filter)
			if err != nil {
				t.Fatal(err)
			}
			if data.Name != "demo" || len(data.EventsAllowed) != 1 || data.EventsAllowed[0].Type != "StasisStart" || data.Key == nil || data.Key.ID != "demo" {
				t.Errorf("decoded filter response = %+v", data)
			}
		})
	}
}

func TestAsteriskPingWireAndStatus(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(request *http.Request) (*http.Response, error) {
				if request.Method != http.MethodGet || request.URL.EscapedPath() != "/ari/asterisk/ping" || request.Body != nil {
					t.Errorf("request = %s %s body=%v", request.Method, request.URL.EscapedPath(), request.Body)
				}
				if status != http.StatusOK {
					return ari23Response(request, status, ""), nil
				}
				return ari23Response(request, status, `{"asterisk_id":"node-1","ping":"pong","timestamp":"2026-09-25T12:00:00Z"}`), nil
			})}})
			response, err := client.Asterisk().Ping(nil)
			if status != http.StatusOK {
				if response != nil || CodeFromError(err) != status {
					t.Fatalf("response=%+v error=%v code=%d", response, err, CodeFromError(err))
				}
				return
			}
			if err != nil || response == nil || response.Ping != "pong" || response.AsteriskID != "node-1" || response.Timestamp == "" {
				t.Fatalf("response=%+v error=%v", response, err)
			}
		})
	}
}
