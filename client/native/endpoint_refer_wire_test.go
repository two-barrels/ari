// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestEndpointReferWire(t *testing.T) {
	falseValue := false
	trueValue := true
	tests := []struct {
		name, path, body string
		query            url.Values
		options          ari.EndpointReferOptions
		call             func(ari.Endpoint, ari.EndpointReferOptions) error
	}{
		{"URI with explicit false and variables", "/ari/endpoints/refer", `{"variables":{"display_name":"A + B","X-Trace":"a/b"}}`, url.Values{"to": {"PJSIP/alice+1"}, "from": {"sip:operator@example.com"}, "refer_to": {"PJSIP/bob?x=1&y=2"}, "to_self": {"false"}},
			ari.EndpointReferOptions{To: "PJSIP/alice+1", From: "sip:operator@example.com", ReferTo: "PJSIP/bob?x=1&y=2", ToSelf: &falseValue, Variables: map[string]string{"display_name": "A + B", "X-Trace": "a/b"}},
			func(e ari.Endpoint, opts ari.EndpointReferOptions) error { return e.Refer(nil, opts) }},
		{"endpoint with explicit true and variables", "/ari/endpoints/PJSIP/alice/refer", `{"variables":{"display_name":"Alice"}}`, url.Values{"from": {"sip:operator@example.com"}, "refer_to": {"PJSIP/bob"}, "to_self": {"true"}},
			ari.EndpointReferOptions{From: "sip:operator@example.com", ReferTo: "PJSIP/bob", ToSelf: &trueValue, Variables: map[string]string{"display_name": "Alice"}},
			func(e ari.Endpoint, opts ari.EndpointReferOptions) error {
				return e.ReferToEndpoint(ari.NewEndpointKey("PJSIP", "alice"), opts)
			}},
		{"endpoint with omitted optional fields", "/ari/endpoints/PJSIP/alice/refer", "", url.Values{"from": {"operator"}, "refer_to": {"PJSIP/bob"}},
			ari.EndpointReferOptions{From: "operator", ReferTo: "PJSIP/bob"},
			func(e ari.Endpoint, opts ari.EndpointReferOptions) error {
				return e.ReferToEndpoint(ari.NewEndpointKey("PJSIP", "alice"), opts)
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != test.path {
					t.Errorf("request = %s %s", req.Method, req.URL.EscapedPath())
				}
				query, err := url.ParseQuery(req.URL.RawQuery)
				if err != nil || !reflect.DeepEqual(query, test.query) {
					t.Errorf("query = %v, want %v: %v", query, test.query, err)
				}
				if test.body == "" {
					if req.Body != nil {
						t.Error("unexpected request body")
					}
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
						t.Errorf("body = %s, want %s", body, test.body)
					}
				}
				return ari23Response(req, http.StatusNoContent, ""), nil
			})}})
			if err := test.call(client.Endpoint(), test.options); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEndpointReferValidationAndStatus(t *testing.T) {
	client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
		return ari23Response(req, http.StatusBadRequest, ""), nil
	})}})
	if err := client.Endpoint().Refer(nil, ari.EndpointReferOptions{From: "operator", ReferTo: "PJSIP/bob"}); err == nil {
		t.Fatal("expected required to error")
	}
	if err := client.Endpoint().ReferToEndpoint(ari.NewEndpointKey("PJSIP", "alice"), ari.EndpointReferOptions{ReferTo: "PJSIP/bob"}); err == nil {
		t.Fatal("expected required from error")
	}
	if err := client.Endpoint().ReferToEndpoint(ari.NewKey(ari.EndpointKey, "invalid"), ari.EndpointReferOptions{From: "operator", ReferTo: "PJSIP/bob"}); err == nil {
		t.Fatal("expected invalid endpoint key error")
	}
	if err := client.Endpoint().Refer(nil, ari.EndpointReferOptions{To: "PJSIP/alice", From: "operator", ReferTo: "PJSIP/bob"}); CodeFromError(err) != http.StatusBadRequest {
		t.Fatalf("error = %v", err)
	}
}
