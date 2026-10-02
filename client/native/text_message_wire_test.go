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

func TestTextMessageWire(t *testing.T) {
	tests := []struct {
		name, path string
		query      url.Values
		body       string
		call       func(ari.TextMessage) error
	}{
		{"endpoint", "/ari/endpoints/PJSIP/alice+1/sendMessage", url.Values{"from": {"sip:operator@example.com"}, "body": {"hello & goodbye"}}, `{"variables":{"X-Trace":"a/b"}}`, func(m ari.TextMessage) error {
			return m.SendWithKey(ari.NewEndpointKey("PJSIP", "alice+1"), "sip:operator@example.com", "hello & goodbye", map[string]string{"X-Trace": "a/b"})
		}},
		{"URI with empty variables", "/ari/endpoints/sendMessage", url.Values{"from": {"sip:operator@example.com"}, "to": {"PJSIP/bob?x=1&y=2"}, "body": {"hello"}}, `{"variables":{}}`, func(m ari.TextMessage) error {
			return m.SendByURIWithKey(ari.NodeKey("demo", "node-1"), "sip:operator@example.com", "PJSIP/bob?x=1&y=2", "hello", nil)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPut || req.URL.EscapedPath() != test.path {
					t.Errorf("request = %s %s", req.Method, req.URL.EscapedPath())
				}
				query, err := url.ParseQuery(req.URL.RawQuery)
				if err != nil || !reflect.DeepEqual(query, test.query) {
					t.Errorf("query=%v want=%v error=%v", query, test.query, err)
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
			if err := test.call(client.TextMessage()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTextMessageStatusAndRequiredFields(t *testing.T) {
	client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
		return ari23Response(req, http.StatusNotFound, ""), nil
	})}})
	m := client.TextMessage()
	if err := m.SendWithKey(ari.NewEndpointKey("PJSIP", "alice"), "operator", "hello", nil); CodeFromError(err) != http.StatusNotFound {
		t.Fatalf("error=%v", err)
	}
	if err := m.SendWithKey(nil, "operator", "hello", nil); err == nil {
		t.Fatal("expected endpoint key error")
	}
	if err := m.SendByURI("", "PJSIP/bob", "hello", nil); err == nil {
		t.Fatal("expected source error")
	}
}
