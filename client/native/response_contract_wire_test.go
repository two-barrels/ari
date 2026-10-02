// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"net/http"
	"strings"
	"testing"

	"github.com/two-barrels/ari/v6"
)

// These fixtures exercise response IDs, empty successes, and ARI error codes
// through public native methods. They do not stand in for a live server.
func TestARIResponseContractWire(t *testing.T) {
	bridgeKey := ari.NewKey(ari.BridgeKey, "bridge-1")
	channelKey := ari.NewKey(ari.ChannelKey, "channel-1")
	fixtures := []struct {
		name, path, body, wantID, wantMessage string
		status, wantCode                      int
		wantError                             bool
		call                                  func(*Client) (string, error)
	}{
		{"bridge create returns assigned ID", "/ari/bridges", `{"id":"assigned-bridge"}`, "assigned-bridge", "", 200, 0, false,
			func(c *Client) (string, error) {
				h, err := c.Bridge().CreateWithoutID(nil, ari.BridgeCreateOptions{Type: "mixing"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"bridge create missing ID", "/ari/bridges", `{}`, "", "", 200, 0, true,
			func(c *Client) (string, error) {
				h, err := c.Bridge().CreateWithoutID(nil, ari.BridgeCreateOptions{Type: "mixing"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"bridge create conflict", "/ari/bridges", `{"message":"bridge already exists"}`, "", "bridge already exists", 409, 409, true,
			func(c *Client) (string, error) {
				h, err := c.Bridge().CreateWithoutID(nil, ari.BridgeCreateOptions{Type: "mixing"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"originate path ID", "/ari/channels/channel-9", `{"id":"channel-9"}`, "channel-9", "", 200, 0, false,
			func(c *Client) (string, error) {
				h, err := c.Channel().OriginateWithID(nil, ari.OriginateRequest{Endpoint: "PJSIP/alice", ChannelID: "channel-9", App: "demo"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"originate mismatched ID", "/ari/channels/channel-9", `{"id":"other-channel"}`, "", "differs from requested", 200, 0, true,
			func(c *Client) (string, error) {
				h, err := c.Channel().OriginateWithID(nil, ari.OriginateRequest{Endpoint: "PJSIP/alice", ChannelID: "channel-9", App: "demo"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"originate invalid options", "/ari/channels/channel-9", `{"message":"invalid endpoint"}`, "", "invalid endpoint", 400, 400, true,
			func(c *Client) (string, error) {
				h, err := c.Channel().OriginateWithID(nil, ari.OriginateRequest{Endpoint: "PJSIP/alice", ChannelID: "channel-9", App: "demo"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"bridge play returns assigned ID", "/ari/bridges/bridge-1/play", `{"id":"assigned-playback"}`, "assigned-playback", "", 200, 0, false,
			func(c *Client) (string, error) {
				h, err := c.Bridge().PlayWithoutID(bridgeKey, ari.BridgePlayOptions{Media: []string{"sound:one"}})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"bridge play missing ID", "/ari/bridges/bridge-1/play", `{}`, "", "", 200, 0, true,
			func(c *Client) (string, error) {
				h, err := c.Bridge().PlayWithoutID(bridgeKey, ari.BridgePlayOptions{Media: []string{"sound:one"}})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"bridge play not found", "/ari/bridges/bridge-1/play", `{"message":"Bridge not found"}`, "", "Bridge not found", 404, 404, true,
			func(c *Client) (string, error) {
				h, err := c.Bridge().PlayWithoutID(bridgeKey, ari.BridgePlayOptions{Media: []string{"sound:one"}})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"channel play returns assigned ID", "/ari/channels/channel-1/play", `{"id":"assigned-playback"}`, "assigned-playback", "", 200, 0, false,
			func(c *Client) (string, error) {
				h, err := c.Channel().PlayWithoutID(channelKey, ari.ChannelPlayOptions{Media: []string{"sound:one"}})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"channel play missing ID", "/ari/channels/channel-1/play", `{}`, "", "", 200, 0, true,
			func(c *Client) (string, error) {
				h, err := c.Channel().PlayWithoutID(channelKey, ari.ChannelPlayOptions{Media: []string{"sound:one"}})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"snoop returns assigned ID", "/ari/channels/channel-1/snoop", `{"id":"assigned-snoop"}`, "assigned-snoop", "", 200, 0, false,
			func(c *Client) (string, error) {
				h, err := c.Channel().SnoopWithoutID(channelKey, &ari.SnoopOptions{App: "demo"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"snoop conflict", "/ari/channels/channel-1/snoop", `{"message":"Channel already exists"}`, "", "Channel already exists", 409, 409, true,
			func(c *Client) (string, error) {
				h, err := c.Channel().SnoopWithoutID(channelKey, &ari.SnoopOptions{App: "demo"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != fixture.path {
					t.Errorf("request=%s %s, want POST %s", req.Method, req.URL.EscapedPath(), fixture.path)
				}
				return ari23Response(req, fixture.status, fixture.body), nil
			})}})
			id, err := fixture.call(client)
			if id != fixture.wantID || (err != nil) != fixture.wantError || CodeFromError(err) != fixture.wantCode {
				t.Fatalf("id=%q error=%v code=%d, want id=%q error=%t code=%d", id, err, CodeFromError(err), fixture.wantID, fixture.wantError, fixture.wantCode)
			}
			if fixture.wantMessage != "" && !strings.Contains(err.Error(), fixture.wantMessage) {
				t.Errorf("error %q omits %q", err, fixture.wantMessage)
			}
		})
	}
}
