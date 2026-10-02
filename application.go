// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package ari

// Application represents a communication path interacting with an Asterisk
// server for application-level resources
type Application interface {
	// List returns the list of applications in Asterisk, optionally using the key for filtering
	List(*Key) ([]*Key, error)

	// Get returns a handle to the application for further interaction
	Get(key *Key) *ApplicationHandle

	// Data returns the applications data
	Data(key *Key) (*ApplicationData, error)

	// Subscribe subscribes the given application to an event source
	// event source may be one of:
	//  - channel:<channelId>
	//  - bridge:<bridgeId>
	//  - endpoint:<tech>/<resource> (e.g. SIP/102)
	//  - deviceState:<deviceName>
	Subscribe(key *Key, eventSource string) error

	// Unsubscribe unsubscribes (removes a subscription to) a given
	// ARI application from the provided event source
	// Equivalent to DELETE /applications/{applicationName}/subscription
	Unsubscribe(key *Key, eventSource string) error

	// FilterEvents updates the allowed and/or disallowed event types.
	// A nil filter clears both lists. Omitted lists remain unchanged.
	FilterEvents(key *Key, filter *ApplicationEventFilter) (*ApplicationData, error)

	// ClaimChannel atomically claims a broadcast channel for this application.
	ClaimChannel(key *Key, channelID string) error
}

// EventTypeFilter names an ARI event type in an application event filter.
type EventTypeFilter struct {
	Type string `json:"type"`
}

// ApplicationEventFilter controls which event types reach an application.
// Use a pointer to an empty slice to clear one list without changing the other.
type ApplicationEventFilter struct {
	Allowed    *[]EventTypeFilter `json:"allowed,omitempty"`
	Disallowed *[]EventTypeFilter `json:"disallowed,omitempty"`
}

// ApplicationData describes the data for a Stasis (Ari) application
type ApplicationData struct {
	// Key is the unique identifier for this application instance in the cluster
	Key *Key `json:"key"`

	BridgeIDs        []string          `json:"bridge_ids"`   // Subscribed BridgeIds
	ChannelIDs       []string          `json:"channel_ids"`  // Subscribed ChannelIds
	DeviceNames      []string          `json:"device_names"` // Subscribed Device names
	EndpointIDs      []string          `json:"endpoint_ids"` // Subscribed Endpoints (tech/resource format)
	Name             string            `json:"name"`         // Name of the application
	EventsAllowed    []EventTypeFilter `json:"events_allowed"`
	EventsDisallowed []EventTypeFilter `json:"events_disallowed"`
}

// ApplicationHandle provides a wrapper to an Application interface for
// operations on a specific application
type ApplicationHandle struct {
	key *Key
	a   Application
}

// NewApplicationHandle creates a new handle to the application name
func NewApplicationHandle(key *Key, app Application) *ApplicationHandle {
	return &ApplicationHandle{
		key: key,
		a:   app,
	}
}

// ID returns the identifier for the application
func (ah *ApplicationHandle) ID() string {
	return ah.key.ID
}

// Key returns the key of the application
func (ah *ApplicationHandle) Key() *Key {
	return ah.key
}

// Data retrives the data for the application
func (ah *ApplicationHandle) Data() (ad *ApplicationData, err error) {
	ad, err = ah.a.Data(ah.key)
	return
}

// Subscribe subscribes the application to an event source
// event source may be one of:
//   - channel:<channelId>
//   - bridge:<bridgeId>
//   - endpoint:<tech>/<resource> (e.g. SIP/102)
//   - deviceState:<deviceName>
func (ah *ApplicationHandle) Subscribe(eventSource string) (err error) {
	err = ah.a.Subscribe(ah.key, eventSource)
	return
}

// Unsubscribe unsubscribes (removes a subscription to) a given
// ARI application from the provided event source
// Equivalent to DELETE /applications/{applicationName}/subscription
func (ah *ApplicationHandle) Unsubscribe(eventSource string) (err error) {
	err = ah.a.Unsubscribe(ah.key, eventSource)
	return
}

// FilterEvents updates this application's event type filters.
func (ah *ApplicationHandle) FilterEvents(filter *ApplicationEventFilter) (*ApplicationData, error) {
	return ah.a.FilterEvents(ah.key, filter)
}

// ClaimChannel claims a broadcast channel for this application.
func (ah *ApplicationHandle) ClaimChannel(channelID string) error {
	return ah.a.ClaimChannel(ah.key, channelID)
}

// Match returns true fo the event matches the application
func (ah *ApplicationHandle) Match(e Event) bool {
	return e.GetApplication() == ah.key.ID
}
