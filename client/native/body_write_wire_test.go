// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/two-barrels/ari/v6"
)

func TestJSONBodyWritesPreserveValues(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		query url.Values
		call  func(*Client) error
	}{
		{"application subscription", "/ari/applications/demo/subscription", url.Values{"eventSource": {"channel:call+leg/2"}}, func(c *Client) error {
			return c.Application().Subscribe(ari.NewKey(ari.ApplicationKey, "demo"), "channel:call+leg/2")
		}},
		{"global variable empty value", "/ari/asterisk/variable", url.Values{"variable": {"CUSTOM+NAME"}, "value": {""}}, func(c *Client) error {
			return c.Asterisk().Variables().Set(ari.NewKey(ari.VariableKey, "CUSTOM+NAME"), "")
		}},
		{"global variable escaped value", "/ari/asterisk/variable", url.Values{"variable": {"CUSTOM+NAME"}, "value": {"a=b&c=d"}}, func(c *Client) error {
			return c.Asterisk().Variables().Set(ari.NewKey(ari.VariableKey, "CUSTOM+NAME"), "a=b&c=d")
		}},
		{"channel dial", "/ari/channels/channel-1/dial", url.Values{"caller": {"call+leg/2"}, "timeout": {"7"}}, func(c *Client) error {
			return c.Channel().Dial(ari.NewKey(ari.ChannelKey, "channel-1"), "call+leg/2", 7*time.Second)
		}},
		{"channel dial zero timeout", "/ari/channels/channel-1/dial", url.Values{"timeout": {"0"}}, func(c *Client) error {
			return c.Channel().Dial(ari.NewKey(ari.ChannelKey, "channel-1"), "", 0)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != tc.path || len(req.URL.Query()) != 0 {
					t.Errorf("request=%s %s?%s, want POST %s without query", req.Method, req.URL.EscapedPath(), req.URL.RawQuery, tc.path)
				}
				assertJSONBodyValues(t, req, tc.query)
				return ari23Response(req, http.StatusNoContent, ""), nil
			})}})
			if err := tc.call(c); err != nil {
				t.Fatal(err)
			}
		})
	}
}
