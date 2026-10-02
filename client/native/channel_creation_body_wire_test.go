// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestChannelCreationJSONBodyWire(t *testing.T) {
	originate := ari.OriginateRequest{
		Endpoint: "PJSIP/alice+1", Extension: "s+1", Context: "support/main", Priority: 2,
		Label: "start+1", App: "demo+app", AppArgs: "one,two", CallerID: "Alice <100>",
		Timeout: 15, ChannelID: "channel-1", OtherChannelID: "channel-2",
		Originator: "parent+1", Formats: "ulaw,slin16", Variables: map[string]string{"STATE": "Ready"},
	}
	created := ari.ChannelCreateRequest{
		Endpoint: "PJSIP/alice+1", App: "demo+app", AppArgs: "one,two", ChannelID: "channel-1",
		OtherChannelID: "channel-2", Originator: "parent+1", Formats: "ulaw,slin16",
		Variables: map[string]string{"STATE": "Ready"},
	}
	tests := []struct {
		name, path string
		body       map[string]any
		call       func(*Client) error
	}{
		{"originate collection", "/ari/channels", map[string]any{
			"endpoint": originate.Endpoint, "extension": originate.Extension, "context": originate.Context,
			"priority": float64(2), "label": originate.Label, "app": originate.App, "appArgs": originate.AppArgs,
			"callerId": originate.CallerID, "timeout": float64(15), "channelId": originate.ChannelID,
			"otherChannelId": originate.OtherChannelID, "originator": originate.Originator,
			"formats": originate.Formats, "variables": map[string]any{"STATE": "Ready"},
		}, func(c *Client) error {
			_, err := c.Channel().Originate(nil, originate)
			return err
		}},
		{"originate path ID", "/ari/channels/channel-1", map[string]any{
			"endpoint": originate.Endpoint, "extension": originate.Extension, "context": originate.Context,
			"priority": float64(2), "label": originate.Label, "app": originate.App, "appArgs": originate.AppArgs,
			"callerId": originate.CallerID, "timeout": float64(15), "channelId": originate.ChannelID,
			"otherChannelId": originate.OtherChannelID, "originator": originate.Originator,
			"formats": originate.Formats, "variables": map[string]any{"STATE": "Ready"},
		}, func(c *Client) error {
			_, err := c.Channel().OriginateWithID(nil, originate)
			return err
		}},
		{"create", "/ari/channels/create", map[string]any{
			"endpoint": created.Endpoint, "app": created.App, "appArgs": created.AppArgs,
			"channelId": created.ChannelID, "otherChannelId": created.OtherChannelID,
			"originator": created.Originator, "formats": created.Formats,
			"variables": map[string]any{"STATE": "Ready"},
		}, func(c *Client) error {
			_, err := c.Channel().Create(nil, created)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != test.path || len(req.URL.Query()) != 0 {
					t.Errorf("request=%s %s?%s, want POST %s without query", req.Method, req.URL.EscapedPath(), req.URL.RawQuery, test.path)
				}
				data, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatal(err)
				}
				var body map[string]any
				if err := json.Unmarshal(data, &body); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(body, test.body) {
					t.Errorf("body=%s, want %#v", data, test.body)
				}
				return ari23Response(req, http.StatusOK, `{"id":"channel-1"}`), nil
			})}})
			if err := test.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}
