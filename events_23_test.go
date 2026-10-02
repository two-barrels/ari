// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package ari

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAsterisk23EventTypesDecode(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want any
		keys []string
	}{
		{"ApplicationRegistered", `{"type":"ApplicationRegistered","application":"demo"}`, &ApplicationRegistered{}, nil},
		{"ApplicationUnregistered", `{"type":"ApplicationUnregistered","application":"demo"}`, &ApplicationUnregistered{}, nil},
		{"CallBroadcast", `{"type":"CallBroadcast","application":"demo","channel":{"id":"channel-1"},"caller":"100","called":"200"}`, &CallBroadcast{}, []string{"channel:channel-1"}},
		{"CallClaimed", `{"type":"CallClaimed","application":"demo","channel":{"id":"channel-1"},"winner_app":"winner"}`, &CallClaimed{}, []string{"channel:channel-1"}},
		{"ChannelToneDetected", `{"type":"ChannelToneDetected","application":"demo","channel":{"id":"channel-1"}}`, &ChannelToneDetected{}, []string{"channel:channel-1"}},
		{"ChannelTransfer", `{"type":"ChannelTransfer","application":"demo","state":"Success","refer_to":{"requested_destination":{"destination":"200"},"destination_channel":{"id":"channel-2"},"connected_channel":{"id":"channel-3"},"bridge":{"id":"bridge-1","channels":["channel-2"]}},"referred_by":{"source_channel":{"id":"channel-1"},"connected_channel":{"id":"channel-3"},"bridge":{"id":"bridge-1"}}}`, &ChannelTransfer{}, []string{"channel:channel-2", "channel:channel-3", "channel:channel-1", "bridge:bridge-1"}},
		{"RESTResponse", `{"type":"RESTResponse","application":"demo","transaction_id":"tx-1","request_id":"req-1","status_code":200,"reason_phrase":"OK","uri":"/channels"}`, &RESTResponse{}, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := DecodeEvent([]byte(test.raw))
			if err != nil {
				t.Fatal(err)
			}
			if reflect.TypeOf(event) != reflect.TypeOf(test.want) {
				t.Fatalf("decoded %T, want %T", event, test.want)
			}
			if event.GetType() != test.name {
				t.Errorf("type = %q", event.GetType())
			}
			var keys []string
			for _, key := range event.Keys() {
				keys = append(keys, key.Kind+":"+key.ID)
			}
			if !reflect.DeepEqual(keys, test.keys) {
				t.Errorf("keys = %v, want %v", keys, test.keys)
			}
			encoded, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.raw), &want); err != nil {
				t.Fatal(err)
			}
			if !containsEventFields(got, want) {
				t.Errorf("round trip lost input fields: %s, want %s", encoded, test.raw)
			}
		})
	}
}

func containsEventFields(got, want any) bool {
	wanted, ok := want.(map[string]any)
	if !ok {
		return reflect.DeepEqual(got, want)
	}
	actual, ok := got.(map[string]any)
	if !ok {
		return false
	}
	for name, value := range wanted {
		if !containsEventFields(actual[name], value) {
			return false
		}
	}
	return true
}

func TestAsterisk23TransferPreservesNestedExtensions(t *testing.T) {
	raw := []byte(`{"type":"ChannelTransfer","application":"demo","refer_to":{"requested_destination":{"destination":"200","future_destination":{"token":7}},"destination_channel":{"id":"channel-2","future_channel":true}},"referred_by":{"source_channel":{"id":"channel-1"}},"future_event":"retained"}`)
	event, err := DecodeEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	transfer, ok := event.(*ChannelTransfer)
	if !ok {
		t.Fatalf("decoded %T", event)
	}
	if transfer.ReferTo.RequestedDestination.Destination != "200" || transfer.ReferredBy.SourceChannel.ID != "channel-1" {
		t.Fatalf("typed transfer fields = %+v", transfer)
	}
	transfer.SetDialog("dialog-1")
	encoded, err := json.Marshal(transfer)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	referTo := got["refer_to"].(map[string]any)
	destination := referTo["requested_destination"].(map[string]any)
	if got["future_event"] != "retained" || got["dialog"] != "dialog-1" ||
		destination["future_destination"].(map[string]any)["token"] != float64(7) ||
		referTo["destination_channel"].(map[string]any)["future_channel"] != true {
		t.Errorf("nested event fields lost: %s", encoded)
	}
}
