// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package main

// Generate event code from Asterisk's events.json Swagger specification.

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var typeMappings = map[string]string{
	"boolean":         "bool",
	"Channel":         "ChannelData",
	"Bridge":          "BridgeData",
	"Playback":        "PlaybackData",
	"LiveRecording":   "LiveRecordingData",
	"StoredRecording": "StoredRecordingData",
	"Endpoint":        "EndpointData",
	"DeviceState":     "DeviceStateData",
	"TextMessage":     "TextMessageData",
	"object":          "any",
}

type spec struct {
	Models map[string]model `json:"models"`
}

type model struct {
	Description string               `json:"description"`
	Properties  map[string]modelProp `json:"properties"`
	SubTypes    []string             `json:"subTypes"`
}

type modelProp struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Required    *bool  `json:"required"`
}

type event struct {
	Name        string
	Event       string
	Description string
	Properties  []prop
}

type prop struct {
	Name        string
	JSONName    string
	Mapping     string
	Type        string
	Description string
}

type generated struct {
	Events []event
	Models []event
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <template> <events.json>", os.Args[0])
	}
	data, err := os.ReadFile(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	result, err := generate(data)
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.ParseFiles(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if err := tmpl.ExecuteTemplate(os.Stdout, "template.tmpl", result); err != nil {
		log.Fatal(err)
	}
}

func generate(data []byte) (generated, error) {
	var source spec
	if err := json.Unmarshal(data, &source); err != nil {
		return generated{}, err
	}
	base, ok := source.Models["Event"]
	if !ok || len(base.SubTypes) == 0 {
		return generated{}, fmt.Errorf("Event.subTypes is empty or missing")
	}
	result := generated{}
	seen := make(map[string]bool)
	dependencies := make(map[string]bool)
	var resolveType func(string) (string, error)
	resolveType = func(name string) (string, error) {
		if strings.HasPrefix(name, "List[") && strings.HasSuffix(name, "]") {
			item, err := resolveType(strings.TrimSuffix(strings.TrimPrefix(name, "List["), "]"))
			return "[]" + item, err
		}
		if mapped, ok := typeMappings[name]; ok {
			return mapped, nil
		}
		switch name {
		case "string", "int", "bool":
			return name, nil
		}
		if _, ok := source.Models[name]; ok {
			dependencies[name] = true
			return normalizeName(name), nil
		}
		return "", fmt.Errorf("unknown event property type %q", name)
	}
	build := func(name string) (event, error) {
		model, ok := source.Models[name]
		if !ok {
			return event{}, fmt.Errorf("missing model %q", name)
		}
		item := event{Name: normalizeName(name), Event: name, Description: cleanDescription(model.Description)}
		for jsonName, property := range model.Properties {
			typ, err := resolveType(property.Type)
			if err != nil {
				return event{}, fmt.Errorf("%s.%s: %w", name, jsonName, err)
			}
			var fieldName string
			for _, part := range strings.Split(jsonName, "_") {
				fieldName += cases.Title(language.English).String(part)
			}
			mapping := "`json:\"" + jsonName + "\"`"
			if property.Required != nil && !*property.Required {
				omit := "omitempty"
				// omitempty does not omit a zero struct, even if it implements
				// MarshalJSON. This made absent optional event payloads appear
				// with fabricated zero-valued resource data on the wire.
				if _, ok := source.Models[property.Type]; ok && typ != "any" {
					omit = "omitzero"
				}
				if strings.HasSuffix(typ, "Data") {
					omit = "omitzero"
				}
				mapping = "`json:\"" + jsonName + "," + omit + "\"`"
			}
			description := cleanDescription(property.Description)
			if description != "" {
				description = "// " + description
			}
			item.Properties = append(item.Properties, prop{
				Name: fieldName, JSONName: jsonName, Type: typ, Mapping: mapping,
				Description: description,
			})
		}
		sort.Slice(item.Properties, func(i, j int) bool { return item.Properties[i].Name < item.Properties[j].Name })
		return item, nil
	}
	for _, name := range base.SubTypes {
		if seen[name] {
			return generated{}, fmt.Errorf("duplicate Event.subTypes entry %q", name)
		}
		seen[name] = true
		item, err := build(name)
		if err != nil {
			return generated{}, err
		}
		result.Events = append(result.Events, item)
	}
	sort.Slice(result.Events, func(i, j int) bool { return result.Events[i].Name < result.Events[j].Name })
	// Keep the formerly exported error model as data for source compatibility.
	if _, ok := source.Models["MissingParams"]; ok {
		dependencies["MissingParams"] = true
	}
	for {
		var pending []string
		for name := range dependencies {
			if !seen[name] {
				pending = append(pending, name)
			}
		}
		if len(pending) == 0 {
			break
		}
		sort.Strings(pending)
		for _, name := range pending {
			seen[name] = true
			item, err := build(name)
			if err != nil {
				return generated{}, err
			}
			result.Models = append(result.Models, item)
		}
	}
	sort.Slice(result.Models, func(i, j int) bool { return result.Models[i].Name < result.Models[j].Name })
	return result, nil
}

func normalizeName(name string) string { return strings.ReplaceAll(name, "Id", "ID") }

func cleanDescription(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\n", ""), "\r", "")
}
