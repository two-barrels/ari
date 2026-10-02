// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package ari

import (
	"encoding/json"
	"testing"
)

// Benchmark the typed decode and raw-field merge used when forwarding an event.
func BenchmarkEventDecodeMarshal(b *testing.B) {
	raw := []byte(`{"type":"StasisStart","application":"demo","asterisk_id":"node-1","args":["first","second"],"channel":{"id":"channel-1","name":"PJSIP/100-00000001","state":"Up","creationtime":"2025-01-01T00:00:00.000+0000","caller":{"name":"Alice","number":"100"},"connected":{"name":"Bob","number":"200"},"future_channel_field":{"codec":"opus","media":{"rx":1234,"tx":5678}}},"future_event_field":{"trace":"abc123","metadata":{"queue":"support","region":"west"}}}`)
	b.SetBytes(int64(len(raw)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		event, err := DecodeEvent(raw)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := json.Marshal(event); err != nil {
			b.Fatal(err)
		}
	}
}
