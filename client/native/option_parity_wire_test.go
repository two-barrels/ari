package native

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/two-barrels/ari/v6"
)

func TestChannelNewOptionsWire(t *testing.T) {
	falseValue, trueValue, zero := false, true, 0
	key := ari.NewKey(ari.ChannelKey, "channel-1")
	tests := []struct {
		name, method, path string
		query              url.Values
		call               func(ari.Channel) error
	}{
		{"variable omitted flag", "POST", "/ari/channels/channel-1/variable", url.Values{"variable": {"STATE"}, "value": {""}}, func(c ari.Channel) error { return c.SetVariable(key, "STATE", "") }},
		{"variable explicit false", "POST", "/ari/channels/channel-1/variable", url.Values{"variable": {"STATE"}, "value": {"Ready"}, "report_events": {"false"}}, func(c ari.Channel) error {
			return c.SetVariableWithOptions(key, "STATE", "Ready", &ari.ChannelVariableSetOptions{ReportEvents: &falseValue})
		}},
		{"variable explicit true", "POST", "/ari/channels/channel-1/variable", url.Values{"variable": {"STATE"}, "value": {"Ready"}, "report_events": {"true"}}, func(c ari.Channel) error {
			return c.SetVariableWithOptions(key, "STATE", "Ready", &ari.ChannelVariableSetOptions{ReportEvents: &trueValue})
		}},
		{"hangup detailed cause", "DELETE", "/ari/channels/channel-1", url.Values{"reason": {"busy"}, "reason_code": {"17"}}, func(c ari.Channel) error {
			return c.HangupWithOptions(key, ari.ChannelHangupOptions{Reason: "busy", ReasonCode: "17"})
		}},
		{"continue label without priority", "POST", "/ari/channels/channel-1/continue", url.Values{"context": {"support"}, "extension": {"s"}, "label": {"start+1"}}, func(c ari.Channel) error {
			return c.ContinueWithOptions(key, ari.ChannelContinueOptions{Context: "support", Extension: "s", Label: "start+1"})
		}},
		{"continue explicit zero priority", "POST", "/ari/channels/channel-1/continue", url.Values{"priority": {"0"}}, func(c ari.Channel) error {
			return c.ContinueWithOptions(key, ari.ChannelContinueOptions{Priority: &zero})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != test.method || req.URL.EscapedPath() != test.path {
					t.Errorf("request = %s %s", req.Method, req.URL.EscapedPath())
				}
				if test.method == http.MethodDelete {
					if !reflect.DeepEqual(req.URL.Query(), test.query) || req.Body != nil {
						t.Errorf("DELETE query=%v body=%v want=%v", req.URL.Query(), req.Body, test.query)
					}
				} else {
					if len(req.URL.Query()) != 0 {
						t.Errorf("unexpected query: %s", req.URL.RawQuery)
					}
					assertJSONBodyValues(t, req, test.query)
				}
				return ari23Response(req, http.StatusNoContent, ""), nil
			})}})
			if err := test.call(client.Channel()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBridgeNewOptionsWire(t *testing.T) {
	falseValue := false
	tests := []struct {
		name, path, body string
		query            url.Values
		call             func(ari.Bridge) error
	}{
		{"create variables", "/ari/bridges/bridge-1", `{"type":"mixing","name":"support","variables":{"STATE":{"value":"Ready"},"ALERT":{"value":"Off","report_events":false}}}`, url.Values{}, func(b ari.Bridge) error {
			_, err := b.CreateWithOptions(ari.NewKey(ari.BridgeKey, "bridge-1"), ari.BridgeCreateOptions{Type: "mixing", Name: "support", Variables: map[string]ari.BridgeCreateVariable{"STATE": {Value: "Ready"}, "ALERT": {Value: "Off", ReportEvents: &falseValue}}})
			return err
		}},
		{"add channel explicit false", "/ari/bridges/bridge-1/addChannel", `{"channel":"channel-1","inhibitConnectedLineUpdates":false}`, url.Values{}, func(b ari.Bridge) error {
			return b.AddChannelWithOptions(ari.NewKey(ari.BridgeKey, "bridge-1"), "channel-1", &ari.BridgeAddChannelOptions{InhibitConnectedLineUpdates: &falseValue})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != test.path {
					t.Errorf("request = %s %s", req.Method, req.URL.EscapedPath())
				}
				if got := req.URL.Query(); !reflect.DeepEqual(got, test.query) {
					t.Errorf("query=%v want=%v", got, test.query)
				}
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatal(err)
				}
				var got, want any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte(test.body), &want); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("body=%s want=%s", body, test.body)
				}
				return ari23Response(req, http.StatusNoContent, ""), nil
			})}})
			if err := test.call(client.Bridge()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBridgeMediaOptionsWire(t *testing.T) {
	zero, skip := 0, 4500
	tests := []struct {
		name, path string
		query      url.Values
		body       string
		call       func(ari.Bridge) error
	}{
		{
			name: "play omitted", path: "/ari/bridges/bridge-1/play/playback-1", query: url.Values{"media": {"sound:one"}},
			call: func(b ari.Bridge) error {
				_, err := b.Play(ari.NewKey(ari.BridgeKey, "bridge-1"), "playback-1", "sound:one")
				return err
			},
		},
		{
			name: "play all options", path: "/ari/bridges/bridge-1/play/playback-1",
			query: url.Values{"media": {"sound:one", "sound:two"}, "announcer_format": {"slin16"}, "lang": {"en+US"}, "offsetms": {"0"}, "skipms": {"4500"}},
			call: func(b ari.Bridge) error {
				_, err := b.PlayWithOptions(ari.NewKey(ari.BridgeKey, "bridge-1"), "playback-1", ari.BridgePlayOptions{Media: []string{"sound:one", "sound:two"}, AnnouncerFormat: "slin16", Lang: "en+US", OffsetMS: &zero, SkipMS: &skip})
				return err
			},
		},
		{
			name: "record all options", path: "/ari/bridges/bridge-1/record",
			query: url.Values{"name": {"recording-1"}, "format": {"wav"}, "recorder_format": {"slin16+stereo"}, "maxDurationSeconds": {"10"}, "maxSilenceSeconds": {"2"}, "ifExists": {"overwrite"}, "beep": {"true"}, "terminateOn": {"#"}},
			call: func(b ari.Bridge) error {
				_, err := b.Record(ari.NewKey(ari.BridgeKey, "bridge-1"), "recording-1", &ari.RecordingOptions{Format: "wav", RecorderFormat: "slin16+stereo", MaxDuration: 10 * time.Second, MaxSilence: 2 * time.Second, Exists: "overwrite", Beep: true, Terminate: "#"})
				return err
			},
		},
		{
			name: "record defaults", path: "/ari/bridges/bridge-1/record",
			query: url.Values{"name": {"recording-2"}, "format": {"wav"}, "maxDurationSeconds": {"0"}, "maxSilenceSeconds": {"0"}, "beep": {"false"}},
			call: func(b ari.Bridge) error {
				_, err := b.Record(ari.NewKey(ari.BridgeKey, "bridge-1"), "recording-2", &ari.RecordingOptions{Format: "wav"})
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != test.path || len(req.URL.Query()) != 0 {
					t.Errorf("request=%s %s?%s, want POST %s without query", req.Method, req.URL.EscapedPath(), req.URL.RawQuery, test.path)
				}
				if test.body == "" {
					assertJSONBodyValues(t, req, test.query)
				} else {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						t.Fatal(err)
					}
					var got, want any
					if err := json.Unmarshal(body, &got); err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal([]byte(test.body), &want); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("body=%s want=%s", body, test.body)
					}
				}
				return ari23Response(req, http.StatusOK, `{}`), nil
			})}})
			if err := test.call(client.Bridge()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestChannelRemainingOptionsWire(t *testing.T) {
	zero, skip := 0, 5000
	tests := []struct {
		name, path string
		query      url.Values
		wantBody   map[string]any
		call       func(ari.Channel) error
	}{
		{
			name: "create variables", path: "/ari/channels/create", query: url.Values{},
			wantBody: map[string]any{"endpoint": "PJSIP/alice", "app": "demo", "channelId": "channel-2", "variables": map[string]any{"STATE": "Ready"}},
			call: func(c ari.Channel) error {
				_, err := c.Create(nil, ari.ChannelCreateRequest{Endpoint: "PJSIP/alice", App: "demo", ChannelID: "channel-2", Variables: map[string]string{"STATE": "Ready"}})
				return err
			},
		},
		{
			name: "play all options", path: "/ari/channels/channel-1/play/playback-1",
			query:    url.Values{},
			wantBody: map[string]any{"media": []any{"sound:one", "sound:two"}, "lang": "en+US", "offsetms": float64(0), "skipms": float64(5000)},
			call: func(c ari.Channel) error {
				_, err := c.PlayWithOptions(ari.NewKey(ari.ChannelKey, "channel-1"), "playback-1", ari.ChannelPlayOptions{Media: []string{"sound:one", "sound:two"}, Lang: "en+US", OffsetMS: &zero, SkipMS: &skip})
				return err
			},
		},
		{
			name: "external transport data", path: "/ari/channels/externalMedia", query: url.Values{},
			wantBody: map[string]any{"channelId": "external-1", "app": "demo", "external_host": "127.0.0.1:5000", "format": "slin16", "transport_data": "a=b&c=d"},
			call: func(c ari.Channel) error {
				_, err := c.ExternalMedia(nil, ari.ExternalMediaOptions{ChannelID: "external-1", App: "demo", ExternalHost: "127.0.0.1:5000", Format: "slin16", TransportData: "a=b&c=d"})
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != test.path || !reflect.DeepEqual(req.URL.Query(), test.query) {
					t.Errorf("request=%s %s?%s, want POST %s?%s", req.Method, req.URL.EscapedPath(), req.URL.RawQuery, test.path, test.query.Encode())
				}
				if test.wantBody == nil {
					if req.Body != nil {
						t.Error("unexpected body")
					}
					return ari23Response(req, http.StatusOK, `{}`), nil
				}
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatal(err)
				}
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatal(err)
				}
				for name, want := range test.wantBody {
					if !reflect.DeepEqual(got[name], want) {
						t.Errorf("body[%s]=%v want=%v", name, got[name], want)
					}
				}
				return ari23Response(req, http.StatusOK, `{}`), nil
			})}})
			if err := test.call(client.Channel()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
