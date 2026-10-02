// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"errors"
	"net/url"
	"strings"

	"github.com/two-barrels/ari/v6"
)

// TextMessage provides the ARI TextMessage accessors for the native client
type TextMessage struct {
	client *Client
}

// Send sends a text message to an endpoint
func (t *TextMessage) Send(from, tech, resource, body string, vars map[string]string) error {
	if from == "" || tech == "" || resource == "" {
		return errors.New("message source and endpoint are required")
	}
	// Construct querystring values
	v := url.Values{}
	v.Set("from", from)
	v.Set("body", body)

	// vars must not be nil, or Ari will reject the request
	if vars == nil {
		vars = map[string]string{}
	}

	data := struct {
		Variables map[string]string `json:"variables"`
	}{
		Variables: vars,
	}

	return t.client.put("/endpoints/"+url.PathEscape(tech)+"/"+url.PathEscape(resource)+"/sendMessage?"+v.Encode(), nil, &data)
}

// SendByURI sends a text message to an endpoint by free-form URI (rather than tech/resource)
func (t *TextMessage) SendByURI(from, to, body string, vars map[string]string) error {
	if from == "" || to == "" {
		return errors.New("message source and destination are required")
	}
	// Construct querystring values
	v := url.Values{}
	v.Set("from", from)
	v.Set("to", to)
	v.Set("body", body)

	// vars must not be nil, or Ari will reject the request
	if vars == nil {
		vars = map[string]string{}
	}

	data := struct {
		Variables map[string]string `json:"variables"`
	}{
		Variables: vars,
	}

	return t.client.put("/endpoints/sendMessage?"+v.Encode(), nil, &data)
}

func (t *TextMessage) SendWithKey(key *ari.Key, from, body string, vars map[string]string) error {
	if key == nil || key.Kind != ari.EndpointKey {
		return errors.New("endpoint key not supplied")
	}
	tech, resource, ok := strings.Cut(key.ID, "/")
	if !ok {
		return errors.New("endpoint key must contain technology and resource")
	}
	return t.Send(from, tech, resource, body, vars)
}

func (t *TextMessage) SendByURIWithKey(_ *ari.Key, from, to, body string, vars map[string]string) error {
	return t.SendByURI(from, to, body, vars)
}
