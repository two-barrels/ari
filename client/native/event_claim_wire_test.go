// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestClaimChannelWireAndStatus(t *testing.T) {
	for _, status := range []int{http.StatusNoContent, http.StatusNotFound, http.StatusConflict} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != "/ari/events/claim" || req.Body != nil {
					t.Errorf("request = %s %s body=%v", req.Method, req.URL.EscapedPath(), req.Body)
				}
				query, err := url.ParseQuery(req.URL.RawQuery)
				want := url.Values{"channelId": {"channel+1"}, "application": {"agent/app"}}
				if err != nil || !reflect.DeepEqual(query, want) {
					t.Errorf("query = %v, want %v: %v", query, want, err)
				}
				return ari23Response(req, status, ""), nil
			})}})
			err := client.Application().ClaimChannel(ari.NewKey(ari.ApplicationKey, "agent/app"), "channel+1")
			if status == http.StatusNoContent && err != nil {
				t.Fatal(err)
			}
			if status != http.StatusNoContent && CodeFromError(err) != status {
				t.Fatalf("error = %v, code=%d", err, CodeFromError(err))
			}
		})
	}
}

func TestClaimChannelRequiresApplicationAndChannel(t *testing.T) {
	client := New(&Options{URL: "http://asterisk.test/ari"})
	if err := client.Application().ClaimChannel(nil, "channel-1"); err == nil {
		t.Fatal("expected application error")
	}
	if err := client.Application().ClaimChannel(ari.NewKey(ari.ApplicationKey, "demo"), ""); err == nil {
		t.Fatal("expected channel ID error")
	}
}
