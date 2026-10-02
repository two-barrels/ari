// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func assertJSONBodyValues(t *testing.T, req *http.Request, want url.Values) {
	t.Helper()
	if req.Body == nil {
		t.Fatal("missing JSON body")
	}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Errorf("body=%s, want fields %v", data, want)
	}
	for name, values := range want {
		if name == "media" || len(values) > 1 {
			items, ok := got[name].([]any)
			if !ok || len(items) != len(values) {
				t.Errorf("body[%s]=%v, want %v", name, got[name], values)
				continue
			}
			for i, value := range values {
				if fmt.Sprint(items[i]) != value {
					t.Errorf("body[%s][%d]=%v, want %s", name, i, items[i], value)
				}
			}
			continue
		}
		if fmt.Sprint(got[name]) != values[0] {
			t.Errorf("body[%s]=%v, want %s", name, got[name], values[0])
		}
	}
}

func TestSimpleCommandBodyWire(t *testing.T) {
	tests := []struct {
		name, method, path string
		query              url.Values
		call               func(*Client) error
	}{
		{"logging", "POST", "/ari/asterisk/logging/log-1", url.Values{"configuration": {"notice,error+debug"}}, func(c *Client) error {
			_, err := c.Asterisk().Logging().Create(ari.NewKey(ari.LoggingKey, "log-1"), "notice,error+debug")
			return err
		}},
		{"device state", "PUT", "/ari/deviceStates/device-1", url.Values{"deviceState": {"INUSE+BUSY"}}, func(c *Client) error {
			return c.DeviceState().Update(ari.NewKey(ari.DeviceStateKey, "device-1"), "INUSE+BUSY")
		}},
		{"mailbox", "PUT", "/ari/mailboxes/mailbox-1", url.Values{"oldMessages": {"2"}, "newMessages": {"3"}}, func(c *Client) error {
			return c.Mailbox().Update(ari.NewKey(ari.MailboxKey, "mailbox-1"), 2, 3)
		}},
		{"playback control", "POST", "/ari/playbacks/playback-1/control", url.Values{"operation": {"forward"}}, func(c *Client) error {
			return c.Playback().Control(ari.NewKey(ari.PlaybackKey, "playback-1"), "forward")
		}},
		{"stored recording copy", "POST", "/ari/recordings/stored/source-1/copy", url.Values{"destinationRecordingName": {"target+1"}}, func(c *Client) error {
			_, err := c.StoredRecording().Copy(ari.NewKey(ari.StoredRecordingKey, "source-1"), "target+1")
			return err
		}},
		{"channel move", "POST", "/ari/channels/channel-1/move", url.Values{"app": {"demo+app"}, "appArgs": {"one,two"}}, func(c *Client) error {
			return c.Channel().Move(ari.NewKey(ari.ChannelKey, "channel-1"), "demo+app", "one,two")
		}},
		{"channel mute", "POST", "/ari/channels/channel-1/mute", url.Values{"direction": {"in"}}, func(c *Client) error {
			return c.Channel().Mute(ari.NewKey(ari.ChannelKey, "channel-1"), ari.DirectionIn)
		}},
		{"channel unmute", "DELETE", "/ari/channels/channel-1/mute", url.Values{"direction": {"out"}}, func(c *Client) error {
			return c.Channel().Unmute(ari.NewKey(ari.ChannelKey, "channel-1"), ari.DirectionOut)
		}},
		{"channel moh", "POST", "/ari/channels/channel-1/moh", url.Values{"mohClass": {"custom+class"}}, func(c *Client) error {
			return c.Channel().MOH(ari.NewKey(ari.ChannelKey, "channel-1"), "custom+class")
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				wantQuery := url.Values{}
				if test.method == http.MethodDelete {
					wantQuery = test.query
				}
				if req.Method != test.method || req.URL.EscapedPath() != test.path || !reflect.DeepEqual(req.URL.Query(), wantQuery) {
					t.Errorf("request=%s %s?%s, want %s %s?%s", req.Method, req.URL.EscapedPath(), req.URL.RawQuery, test.method, test.path, wantQuery.Encode())
				}
				if test.method == http.MethodDelete {
					if req.Body != nil {
						t.Error("unexpected DELETE body")
					}
				} else {
					assertJSONBodyValues(t, req, test.query)
				}
				return ari23Response(req, http.StatusOK, `{"name":"target+1"}`), nil
			})}})
			if err := test.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}
