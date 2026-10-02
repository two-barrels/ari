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

func TestChannelUserEventBodyWire(t *testing.T) {
	client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.EscapedPath() != "/ari/events/user/customer%2Falert" || len(req.URL.Query()) != 0 {
			t.Errorf("request=%s %s?%s", req.Method, req.URL.EscapedPath(), req.URL.RawQuery)
		}
		data, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		var body map[string]any
		if err := json.Unmarshal(data, &body); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(body, map[string]any{"application": "demo+app", "source": "channel:channel+1", "variables": map[string]any{"ticket": "42"}}) {
			t.Errorf("body=%s", data)
		}
		return ari23Response(req, http.StatusNoContent, ""), nil
	})}})
	key := &ari.Key{App: "demo+app", Kind: ari.ChannelKey, ID: "channel+1"}
	returnErr := client.Channel().UserEvent(key, &ari.ChannelUserevent{Eventname: "customer/alert", Userevent: map[string]any{"ticket": "42"}})
	if returnErr != nil {
		t.Fatal(returnErr)
	}
}
