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

func TestChannelARI23Wire(t *testing.T) {
	report := false
	assignments := map[string]ari.VariableAssignment{
		"STATE": {Value: "Ready"},
		"LEVEL": {Value: "Gold", ReportEvents: &report},
	}
	tests := []struct {
		name, method, suffix, body, response string
		query                                url.Values
		call                                 func(ari.Channel, *ari.Key) error
	}{
		{"get variables", "GET", "/variables", "", `{"variables":{"STATE":"Ready","LEVEL":"Gold"}}`, url.Values{"variables": {"STATE", "LEVEL"}}, func(c ari.Channel, key *ari.Key) error {
			values, err := c.GetVariables(key, "STATE", "LEVEL")
			if !reflect.DeepEqual(values, map[string]any{"STATE": "Ready", "LEVEL": "Gold"}) {
				t.Errorf("variables = %#v", values)
			}
			return err
		}},
		{"set variables", "POST", "/variables", `{"variables":{"STATE":"Ready","LEVEL":{"value":"Gold","report_events":false}}}`, "", url.Values{}, func(c ari.Channel, key *ari.Key) error {
			return c.SetVariables(key, assignments)
		}},
		{"redirect", "POST", "/redirect", "", "", url.Values{"endpoint": {"PJSIP/100+1"}}, func(c ari.Channel, key *ari.Key) error {
			return c.Redirect(key, "PJSIP/100+1")
		}},
		{"progress", "POST", "/progress", "", "", url.Values{}, func(c ari.Channel, key *ari.Key) error {
			return c.Progress(key)
		}},
		{"transfer progress", "POST", "/transfer_progress", "", "", url.Values{"states": {"channel_progress"}}, func(c ari.Channel, key *ari.Key) error {
			return c.TransferProgress(key, "channel_progress")
		}},
		{"rtp statistics", "GET", "/rtp_statistics", "", `{"txcount":20,"rxcount":19,"txjitter":0.25,"remote_maxjitter":0.5,"rtt":2.5,"local_ssrc":42,"channel_uniqueid":"channel-1"}`, url.Values{}, func(c ari.Channel, key *ari.Key) error {
			stats, err := c.RTPStatistics(key)
			if err == nil && (stats.TxCount != 20 || stats.RxCount != 19 || stats.TxJitter != 0.25 || stats.RemoteMaxJitter != 0.5 || stats.RTT != 2.5 || stats.LocalSSRC != 42 || stats.ChannelUniqueID != "channel-1") {
				t.Errorf("stats = %+v", stats)
			}
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != test.method || req.URL.EscapedPath() != "/ari/channels/channel-1"+test.suffix {
					t.Errorf("request = %s %s", req.Method, req.URL.EscapedPath())
				}
				query, err := url.ParseQuery(req.URL.RawQuery)
				if err != nil || !reflect.DeepEqual(query, test.query) {
					t.Errorf("query = %v, want %v: %v", query, test.query, err)
				}
				if test.body != "" {
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
				} else if req.Body != nil {
					t.Errorf("unexpected body")
				}
				return ari23Response(req, http.StatusOK, test.response), nil
			})}})
			if err := test.call(client.Channel(), ari.NewKey(ari.ChannelKey, "channel-1")); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestChannelRTPStatisticsStatusAndRequiredFields(t *testing.T) {
	client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
		return ari23Response(req, http.StatusNotFound, ""), nil
	})}})
	channel := client.Channel()
	key := ari.NewKey(ari.ChannelKey, "channel-1")
	if stats, err := channel.RTPStatistics(key); stats != nil || CodeFromError(err) != http.StatusNotFound {
		t.Fatalf("stats=%+v error=%v", stats, err)
	}
	if _, err := channel.GetVariables(key); err == nil {
		t.Fatal("expected required variable names error")
	}
	if err := channel.Redirect(key, ""); err == nil {
		t.Fatal("expected required endpoint error")
	}
	if err := channel.TransferProgress(key, ""); err == nil {
		t.Fatal("expected required state error")
	}
}
