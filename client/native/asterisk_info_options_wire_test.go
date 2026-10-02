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

func TestAsteriskInfoOnlyWire(t *testing.T) {
	for _, test := range []struct {
		name  string
		only  string
		query url.Values
	}{
		{name: "unfiltered", query: url.Values{}},
		{name: "filtered", only: "build,system+status", query: url.Values{"only": {"build,system+status"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet || req.URL.EscapedPath() != "/ari/asterisk/info" || !reflect.DeepEqual(req.URL.Query(), test.query) {
					t.Errorf("request=%s %s?%s, want GET /ari/asterisk/info?%s", req.Method, req.URL.EscapedPath(), req.URL.RawQuery, test.query.Encode())
				}
				return ari23Response(req, http.StatusOK, `{"system":{"version":"23.0.0"}}`), nil
			})}})
			var info *ari.AsteriskInfo
			var err error
			if test.only == "" {
				info, err = client.Asterisk().Info(nil)
			} else {
				info, err = client.Asterisk().InfoWithOptions(nil, ari.AsteriskInfoOptions{Only: test.only})
			}
			if err != nil || info == nil || info.SystemInfo.Version != "23.0.0" {
				t.Fatalf("info=%+v error=%v", info, err)
			}
		})
	}
}
