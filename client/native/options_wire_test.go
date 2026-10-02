package native

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/two-barrels/ari/v6"
)

type recordingTransport struct {
	method string
	path   string
	query  string
	body   []byte
}

func (transport *recordingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.method = request.Method
	transport.path = request.URL.EscapedPath()
	transport.query = request.URL.RawQuery
	if request.Body != nil {
		var err error
		transport.body, err = io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"id":"created-1"}`)),
		Request:    request,
	}, nil
}

func TestApplicationUnsubscribeEncodesEventSource(t *testing.T) {
	transport := &recordingTransport{}
	client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: transport}})
	eventSource := "channel:call+leg"
	if err := client.Application().Unsubscribe(ari.NewKey(ari.ApplicationKey, "demo"), eventSource); err != nil {
		t.Fatal(err)
	}
	query, err := url.ParseQuery(transport.query)
	if err != nil {
		t.Fatal(err)
	}
	if transport.method != http.MethodDelete || transport.path != "/ari/applications/demo/subscription" || query.Get("eventSource") != eventSource {
		t.Errorf("request = %s %s?%s; decoded eventSource = %q", transport.method, transport.path, transport.query, query.Get("eventSource"))
	}
}

func TestVariableGetEncodesName(t *testing.T) {
	for _, test := range []struct {
		name string
		path string
		call func(*Client) error
	}{
		{
			name: "channel", path: "/ari/channels/channel-1/variable",
			call: func(c *Client) error {
				_, err := c.Channel().GetVariable(ari.NewKey(ari.ChannelKey, "channel-1"), "CUSTOM+NAME")
				return err
			},
		},
		{
			name: "global", path: "/ari/asterisk/variable",
			call: func(c *Client) error {
				_, err := c.Asterisk().Variables().Get(ari.NewKey(ari.VariableKey, "CUSTOM+NAME"))
				return err
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			transport := &recordingTransport{}
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: transport}})
			if err := test.call(client); err != nil {
				t.Fatal(err)
			}
			query, err := url.ParseQuery(transport.query)
			if err != nil {
				t.Fatal(err)
			}
			if transport.method != http.MethodGet || transport.path != test.path || query.Get("variable") != "CUSTOM+NAME" {
				t.Errorf("request = %s %s?%s; decoded variable = %q", transport.method, transport.path, transport.query, query.Get("variable"))
			}
		})
	}
}

func TestNativeOptionsReachHTTPWire(t *testing.T) {
	tests := []struct {
		name string
		path string
		call func(*Client) error
		want string
	}{
		{
			name: "bridge add channel", path: "/ari/bridges/bridge-1/addChannel",
			call: func(c *Client) error {
				return c.Bridge().AddChannelWithOptions(ari.NewKey(ari.BridgeKey, "bridge-1"), "channel-1", &ari.BridgeAddChannelOptions{
					Role: "caller", AbsorbDTMF: true, Mute: true,
				})
			},
			want: `{"channel":"channel-1","role":"caller","absorbDTMF":true,"mute":true}`,
		},
		{
			name: "external media", path: "/ari/channels/externalMedia",
			call: func(c *Client) error {
				_, err := c.Channel().ExternalMedia(nil, ari.ExternalMediaOptions{
					ChannelID: "external-1", App: "demo", ExternalHost: "127.0.0.1:5000",
					Encapsulation: "rtp", Transport: "udp", ConnectionType: "client",
					Format: "slin16", Direction: "both", Data: "stream-1",
					Variables: map[string]string{"ticket": "42"},
				})
				return err
			},
			want: `{"channelId":"external-1","app":"demo","external_host":"127.0.0.1:5000","encapsulation":"rtp","transport":"udp","connection_type":"client","format":"slin16","direction":"both","data":"stream-1","variables":{"ticket":"42"}}`,
		},
		{
			name: "channel record", path: "/ari/channels/channel-1/record",
			call: func(c *Client) error {
				_, err := c.Channel().Record(ari.NewKey(ari.ChannelKey, "channel-1"), "recording-1", &ari.RecordingOptions{
					Format: "wav", MaxDuration: 10 * time.Second, MaxSilence: 2 * time.Second,
					Exists: "overwrite", Beep: true, Terminate: "#",
				})
				return err
			},
			want: `{"name":"recording-1","format":"wav","maxDurationSeconds":10,"maxSilenceSeconds":2,"ifExists":"overwrite","beep":true,"terminateOn":"#"}`,
		},
		{
			name: "channel DTMF timings", path: "/ari/channels/channel-1/dtmf",
			call: func(c *Client) error {
				return c.Channel().SendDTMF(ari.NewKey(ari.ChannelKey, "channel-1"), "12#", &ari.DTMFOptions{
					Before: 50 * time.Millisecond, Between: 75 * time.Millisecond,
					Duration: 120 * time.Millisecond, After: 25 * time.Millisecond,
				})
			},
			want: `{"dtmf":"12#","before":50,"between":75,"duration":120,"after":25}`,
		},
		{
			name: "channel DTMF defaults", path: "/ari/channels/channel-1/dtmf",
			call: func(c *Client) error {
				return c.Channel().SendDTMF(ari.NewKey(ari.ChannelKey, "channel-1"), "5", nil)
			},
			want: `{"dtmf":"5","between":100,"duration":100}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := &recordingTransport{}
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: transport}})
			if err := test.call(client); err != nil {
				t.Fatal(err)
			}
			if transport.method != http.MethodPost || transport.path != test.path {
				t.Errorf("request = %s %s, want POST %s", transport.method, transport.path, test.path)
			}
			var got, want any
			if err := json.Unmarshal(transport.body, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("request body = %s, want %s", transport.body, test.want)
			}
		})
	}
}
