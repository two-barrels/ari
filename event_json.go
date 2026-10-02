// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package ari

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

// CloneEvent makes an independent top-level event value for dialog routing.
// Nested payload values are shared and should be treated as immutable. The
// event metadata and raw JSON are copied, so SetDialog only affects the clone.
func CloneEvent(event Event) (Event, error) {
	if event == nil {
		return nil, errors.New("cannot clone nil event")
	}
	value := reflect.ValueOf(event)
	if value.Kind() == reflect.Pointer && !value.IsNil() && value.Elem().Kind() == reflect.Struct {
		copyValue := reflect.New(value.Elem().Type())
		copyValue.Elem().Set(value.Elem())
		if copied, ok := copyValue.Interface().(Event); ok {
			if data, ok := copied.(interface{ eventData() *EventData }); ok {
				metadata := data.eventData()
				metadata.raw = append([]byte(nil), metadata.raw...)
			}
			return copied, nil
		}
	}
	// Preserve the previous JSON copy behavior for non-pointer event values.
	encoded, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return DecodeEvent(encoded)
}

func (e *EventData) setRaw(data []byte) {
	e.raw = append(e.raw[:0], data...)
}

// mergeEventJSON keeps fields from Asterisk that the typed model cannot yet
// express. Typed values take precedence when an application changes them.
func mergeEventJSON(original, typed []byte) ([]byte, error) {
	if len(original) == 0 {
		return typed, nil
	}
	var oldFields, newFields map[string]json.RawMessage
	if err := json.Unmarshal(original, &oldFields); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(typed, &newFields); err != nil {
		return nil, err
	}
	for name, value := range newFields {
		prior, present := oldFields[name]
		if present && string(value) == "null" && string(prior) != "null" {
			continue
		}
		if present && len(value) > 0 && value[0] == '{' && len(prior) > 0 && prior[0] == '{' {
			merged, err := mergeEventJSON(prior, value)
			if err != nil {
				return nil, err
			}
			oldFields[name] = merged
			continue
		}
		oldFields[name] = value
	}
	return json.Marshal(oldFields)
}

// UnknownEvent retains newer ARI event types until a typed model is available.
// Application-wide subscribers can still receive these events without an update
// to the generated event catalogue.
type UnknownEvent struct {
	EventData
}

func (e *UnknownEvent) UnmarshalJSON(data []byte) error {
	type alias UnknownEvent
	if err := json.Unmarshal(data, (*alias)(e)); err != nil {
		return err
	}
	e.EventData.setRaw(data)
	return nil
}

func (e *UnknownEvent) MarshalJSON() ([]byte, error) {
	typed, err := json.Marshal(e.EventData)
	if err != nil {
		return nil, err
	}
	return mergeEventJSON(e.EventData.raw, typed)
}

// Keys finds common ARI resource objects in an event unknown to this version.
// Events with no recognized resource remain available to all-event subscribers.
func (e *UnknownEvent) Keys() (keys Keys) {
	var body map[string]any
	if err := json.Unmarshal(e.EventData.raw, &body); err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var walk func(map[string]any)
	walk = func(object map[string]any) {
		for field, value := range object {
			child, ok := value.(map[string]any)
			if !ok {
				continue
			}
			name := strings.ToLower(field)
			kind := ""
			id, _ := child["id"].(string)
			switch {
			case strings.Contains(name, "channel"):
				kind = ChannelKey
			case strings.Contains(name, "bridge"):
				kind = BridgeKey
			case strings.Contains(name, "playback"):
				kind = PlaybackKey
			case strings.Contains(name, "endpoint"):
				kind = EndpointKey
				tech, _ := child["technology"].(string)
				resource, _ := child["resource"].(string)
				id = endpointKeyID(tech, resource)
			}
			if kind != "" && id != "" && !seen[kind+":"+id] {
				keys = append(keys, e.Key(kind, id))
				seen[kind+":"+id] = true
			}
			walk(child)
		}
	}
	walk(body)
	return keys
}
