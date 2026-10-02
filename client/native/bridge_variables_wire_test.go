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

func TestBridgeVariablesWire(t *testing.T) {
	report := false
	values := map[string]ari.BridgeVariableAssignment{
		"BRIDGE_STATE":  {Value: "Waiting"},
		"SUPPORT_LEVEL": {Value: "Premium", ReportEvents: &report},
	}
	tests := []struct {
		name     string
		method   string
		path     string
		query    url.Values
		body     string
		response string
		call     func(ari.Bridge, *ari.Key) error
	}{
		{"get one", http.MethodGet, "/ari/bridges/bridge-1/variable", url.Values{"variable": {"BRIDGE+STATE"}}, "", `{"value":"Waiting"}`, func(bridge ari.Bridge, key *ari.Key) error {
			value, err := bridge.GetVariable(key, "BRIDGE+STATE")
			if value != "Waiting" {
				t.Errorf("variable = %q", value)
			}
			return err
		}},
		{"set one explicit false", http.MethodPost, "/ari/bridges/bridge-1/variable", url.Values{"variable": {"BRIDGE+STATE"}, "value": {""}, "report_events": {"false"}}, "", "", func(bridge ari.Bridge, key *ari.Key) error {
			return bridge.SetVariable(key, "BRIDGE+STATE", "", &report)
		}},
		{"get many", http.MethodGet, "/ari/bridges/bridge-1/variables", url.Values{"variables": {"BRIDGE+STATE", "SUPPORT_LEVEL"}}, "", `{"variables":{"BRIDGE+STATE":"Waiting","SUPPORT_LEVEL":"Premium"}}`, func(bridge ari.Bridge, key *ari.Key) error {
			got, err := bridge.GetVariables(key, "BRIDGE+STATE", "SUPPORT_LEVEL")
			if !reflect.DeepEqual(got, map[string]any{"BRIDGE+STATE": "Waiting", "SUPPORT_LEVEL": "Premium"}) {
				t.Errorf("variables = %+v", got)
			}
			return err
		}},
		{"set many", http.MethodPost, "/ari/bridges/bridge-1/variables", url.Values{}, `{"variables":{"BRIDGE_STATE":"Waiting","SUPPORT_LEVEL":{"value":"Premium","report_events":false}}}`, "", func(bridge ari.Bridge, key *ari.Key) error {
			return bridge.SetVariables(key, values)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(request *http.Request) (*http.Response, error) {
				if request.Method != test.method || request.URL.EscapedPath() != test.path {
					t.Errorf("request = %s %s", request.Method, request.URL.EscapedPath())
				}
				if got, err := url.ParseQuery(request.URL.RawQuery); err != nil || !reflect.DeepEqual(got, test.query) {
					t.Errorf("query = %v, want %v, error = %v", got, test.query, err)
				}
				if test.body != "" {
					body, err := io.ReadAll(request.Body)
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
				} else if request.Body != nil {
					t.Errorf("unexpected request body for %s", test.name)
				}
				return ari23Response(request, http.StatusOK, test.response), nil
			})}})
			if err := test.call(client.Bridge(), ari.NewKey(ari.BridgeKey, "bridge-1")); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBridgeVariablesKeepHTTPStatus(t *testing.T) {
	client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(request *http.Request) (*http.Response, error) {
		return ari23Response(request, http.StatusConflict, ""), nil
	})}})
	key := ari.NewKey(ari.BridgeKey, "bridge-1")
	_, err := client.Bridge().GetVariable(key, "BRIDGE_STATE")
	if CodeFromError(err) != http.StatusConflict {
		t.Errorf("single read error = %v, code = %d", err, CodeFromError(err))
	}
	_, err = client.Bridge().GetVariables(key, "BRIDGE_STATE")
	if CodeFromError(err) != http.StatusConflict {
		t.Errorf("bulk read error = %v, code = %d", err, CodeFromError(err))
	}
	if err := client.Bridge().SetVariable(key, "BRIDGE_STATE", "Waiting", nil); CodeFromError(err) != http.StatusConflict {
		t.Errorf("single write error = %v, code = %d", err, CodeFromError(err))
	}
	if err := client.Bridge().SetVariables(key, map[string]ari.BridgeVariableAssignment{"BRIDGE_STATE": {Value: "Waiting"}}); CodeFromError(err) != http.StatusConflict {
		t.Errorf("bulk write error = %v, code = %d", err, CodeFromError(err))
	}
}
