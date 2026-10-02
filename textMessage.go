// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package ari

// TextMessage needs some verbiage here
type TextMessage interface {
	// Send() sends a text message to an endpoint
	Send(from, tech, resource, body string, vars map[string]string) error

	// SendByURI sends a text message to an endpoint by free-form URI
	SendByURI(from, to, body string, vars map[string]string) error

	// SendWithKey sends to the endpoint identified by key. The key also selects
	// an Asterisk node when used through a proxy.
	SendWithKey(key *Key, from, body string, vars map[string]string) error

	// SendByURIWithKey sends to a URI on the node selected by referenceKey.
	SendByURIWithKey(referenceKey *Key, from, to, body string, vars map[string]string) error
}

// TextMessageData describes text message
type TextMessageData struct {
	// Key is the cluster-unique identifier for this text message
	Key *Key `json:"key"`

	Body      string                `json:"body"` // The body (text) of the message
	From      string                `json:"from"` // Technology-specific source URI
	To        string                `json:"to"`   // Technology-specific destination URI
	Variables []TextMessageVariable `json:"variables,omitempty"`
}

// TextMessageVariable describes a key-value pair (associated with a text message)
type TextMessageVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
