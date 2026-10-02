// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestBridgeMOHAndRemoveChannelBodyWire(t *testing.T) {
	tests := []struct {
		name, path string
		query      url.Values
		call       func(ari.Bridge) error
	}{
		{"moh class", "/ari/bridges/bridge-1/moh", url.Values{"mohClass": {"custom+class"}}, func(b ari.Bridge) error {
			return b.MOH(ari.NewKey(ari.BridgeKey, "bridge-1"), "custom+class")
		}},
		{"moh omitted", "/ari/bridges/bridge-1/moh", url.Values{}, func(b ari.Bridge) error {
			return b.MOH(ari.NewKey(ari.BridgeKey, "bridge-1"), "")
		}},
		{"remove channel", "/ari/bridges/bridge-1/removeChannel", url.Values{"channel": {"channel+1/2"}}, func(b ari.Bridge) error {
			return b.RemoveChannel(ari.NewKey(ari.BridgeKey, "bridge-1"), "channel+1/2")
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != test.path || len(req.URL.Query()) != 0 {
					t.Errorf("request=%s %s?%s, want POST %s", req.Method, req.URL.EscapedPath(), req.URL.RawQuery, test.path)
				}
				if len(test.query) == 0 {
					assertJSONBodyValues(t, req, url.Values{"mohClass": {""}})
				} else {
					assertJSONBodyValues(t, req, test.query)
				}
				return ari23Response(req, http.StatusNoContent, ""), nil
			})}})
			if err := test.call(client.Bridge()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
