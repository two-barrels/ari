package ari

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDecodeEventRetainsUnknownFieldsOnTypedEvent(t *testing.T) {
	raw := []byte(`{"type":"StasisStart","application":"demo","asterisk_id":"node-1","args":["first"],"channel":{"id":"channel-1","creationtime":"2025-01-01T00:00:00.000+0000","future_channel_field":{"nested":[1,"two"]}},"future_event_field":{"enabled":true}}`)
	event, err := DecodeEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	start, ok := event.(*StasisStart)
	if !ok {
		t.Fatalf("decoded event type = %T, want *StasisStart", event)
	}
	start.SetDialog("dialog-1")
	encoded, err := json.Marshal(start)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	if string(result["future_event_field"]) != `{"enabled":true}` {
		t.Errorf("event extension lost: %s", result["future_event_field"])
	}
	if string(result["dialog"]) != `"dialog-1"` {
		t.Errorf("dialog update lost: %s", result["dialog"])
	}
	var channel map[string]json.RawMessage
	if err := json.Unmarshal(result["channel"], &channel); err != nil {
		t.Fatal(err)
	}
	if string(channel["future_channel_field"]) != `{"nested":[1,"two"]}` {
		t.Errorf("channel extension lost: %s", channel["future_channel_field"])
	}
}

func TestStasisStartOmitsAbsentReplaceChannelAcrossEventHops(t *testing.T) {
	raw := []byte(`{"type":"StasisStart","application":"ari-testing","asterisk_id":"node-1","args":[],"channel":{"id":"test-channel","creationtime":"2026-09-28T16:14:28.197+0000"}}`)
	event, err := DecodeEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	first, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	forwarded, err := DecodeEvent(first)
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(forwarded)
	if err != nil {
		t.Fatal(err)
	}
	var firstFields, secondFields map[string]any
	if err := json.Unmarshal(first, &firstFields); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(second, &secondFields); err != nil {
		t.Fatal(err)
	}
	if _, ok := firstFields["replace_channel"]; ok {
		t.Errorf("fabricated replace_channel in first hop: %s", first)
	}
	if !reflect.DeepEqual(firstFields, secondFields) {
		t.Errorf("event changed on second hop: first=%s second=%s", first, second)
	}
}

func TestDecodeEventAcceptsNewerEventType(t *testing.T) {
	raw := []byte(`{"type":"FutureChannelEvent","application":"demo","asterisk_id":"node-1","channel":{"id":"channel-1"},"new_detail":{"value":42}}`)
	event, err := DecodeEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := event.(*UnknownEvent); !ok {
		t.Fatalf("decoded event type = %T, want *UnknownEvent", event)
	}
	if keys := event.Keys(); len(keys) != 1 || keys[0].Kind != ChannelKey || keys[0].ID != "channel-1" {
		t.Errorf("unknown event resource keys = %v", keys)
	}
	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	if string(result["new_detail"]) != `{"value":42}` {
		t.Errorf("new event details lost: %s", result["new_detail"])
	}
}

func TestCloneEventKeepsDialogAndRawFieldsIndependent(t *testing.T) {
	raw := []byte(`{"type":"ChannelTransfer","application":"demo","refer_to":{"requested_destination":{"destination":"200","future":true}},"referred_by":{"source_channel":{"id":"channel-1"}}}`)
	original, err := DecodeEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	clone, err := CloneEvent(original)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := clone.(*ChannelTransfer); !ok {
		t.Fatalf("clone has type %T", clone)
	}
	clone.SetDialog("dialog-1")
	if original.GetDialog() != "" || clone.GetDialog() != "dialog-1" {
		t.Fatalf("dialogs: original=%q clone=%q", original.GetDialog(), clone.GetDialog())
	}
	encoded, err := json.Marshal(clone)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["refer_to"].(map[string]any)["requested_destination"].(map[string]any)["future"] != true {
		t.Errorf("clone lost unknown nested field: %s", encoded)
	}
	clone.(*ChannelTransfer).EventData.raw[0] = ' '
	if original.(*ChannelTransfer).EventData.raw[0] != '{' {
		t.Error("clone changed original raw JSON")
	}
}

func TestCloneUnknownEventKeepsFutureFields(t *testing.T) {
	original, err := DecodeEvent([]byte(`{"type":"FutureEvent","application":"demo","future":{"nested":true}}`))
	if err != nil {
		t.Fatal(err)
	}
	clone, err := CloneEvent(original)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := clone.(*UnknownEvent); !ok {
		t.Fatalf("clone has type %T", clone)
	}
	clone.SetDialog("dialog-1")
	encoded, err := json.Marshal(clone)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["dialog"] != "dialog-1" || fields["future"].(map[string]any)["nested"] != true || original.GetDialog() != "" {
		t.Errorf("unknown event clone = %s; original dialog = %q", encoded, original.GetDialog())
	}
}
