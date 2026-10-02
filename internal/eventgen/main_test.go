// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"text/template"
)

func TestPinnedEventSpecGeneratesDeclaredEventsOnly(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("json", "events-23.json"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := generate(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) != 45 {
		t.Fatalf("generated %d events, want 45", len(result.Events))
	}
	for _, name := range []string{"ApplicationRegistered", "ApplicationUnregistered", "CallBroadcast", "CallClaimed", "ChannelToneDetected", "ChannelTransfer", "RESTResponse"} {
		if !slices.ContainsFunc(result.Events, func(e event) bool { return e.Name == name }) {
			t.Errorf("missing event %s", name)
		}
	}
	for _, name := range []string{"ContactInfo", "MissingParams", "Peer", "ReferTo", "ReferredBy", "RequiredDestination", "AdditionalParam"} {
		if slices.ContainsFunc(result.Events, func(e event) bool { return e.Name == name }) {
			t.Errorf("non-event %s in event switch", name)
		}
		if !slices.ContainsFunc(result.Models, func(e event) bool { return e.Name == name }) {
			t.Errorf("missing data model %s", name)
		}
	}
	if !slices.ContainsFunc(result.Events, func(e event) bool { return e.Name == "ChannelCallerID" && e.Event == "ChannelCallerId" }) {
		t.Error("ChannelCallerId compatibility mapping lost")
	}
	tmpl, err := template.ParseFiles("template.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "template.tmpl", result); err != nil {
		t.Fatal(err)
	}
	formatted, err := format.Source(rendered.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	checkedIn, err := os.ReadFile(filepath.Join("..", "..", "events_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(formatted, checkedIn) {
		t.Error("events_gen.go differs from pinned generated output; run make events")
	}
}
