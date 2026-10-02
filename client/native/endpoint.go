// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/two-barrels/ari/v6"
)

func endpointReferQuery(opts ari.EndpointReferOptions, requireTo bool) (url.Values, error) {
	if opts.From == "" || opts.ReferTo == "" || requireTo && opts.To == "" {
		return nil, errors.New("required REFER parameter not supplied")
	}
	query := url.Values{"from": {opts.From}, "refer_to": {opts.ReferTo}}
	if requireTo {
		query.Set("to", opts.To)
	}
	if opts.ToSelf != nil {
		query.Set("to_self", strconv.FormatBool(*opts.ToSelf))
	}
	return query, nil
}

func endpointReferBody(variables map[string]string) any {
	if variables == nil {
		return nil
	}
	return &struct {
		Variables map[string]string `json:"variables"`
	}{Variables: variables}
}

// Refer sends a REFER to the endpoint or technology URI named by opts.To.
func (e *Endpoint) Refer(_ *ari.Key, opts ari.EndpointReferOptions) error {
	query, err := endpointReferQuery(opts, true)
	if err != nil {
		return err
	}
	return e.client.post("/endpoints/refer?"+query.Encode(), nil, endpointReferBody(opts.Variables))
}

// ReferToEndpoint sends a REFER to the endpoint identified by key.
func (e *Endpoint) ReferToEndpoint(key *ari.Key, opts ari.EndpointReferOptions) error {
	if key == nil || key.Kind != ari.EndpointKey {
		return errors.New("endpoint key not supplied")
	}
	tech, resource, ok := strings.Cut(key.ID, "/")
	if !ok || tech == "" || resource == "" {
		return errors.New("endpoint key must contain technology and resource")
	}
	query, err := endpointReferQuery(opts, false)
	if err != nil {
		return err
	}
	return e.client.post("/endpoints/"+url.PathEscape(tech)+"/"+url.PathEscape(resource)+"/refer?"+query.Encode(), nil, endpointReferBody(opts.Variables))
}

// Endpoint provides the ARI Endpoint accessors for the native client
type Endpoint struct {
	client *Client
}

// Get gets a lazy handle for the endpoint entity
func (e *Endpoint) Get(key *ari.Key) *ari.EndpointHandle {
	return ari.NewEndpointHandle(e.client.stamp(key), e)
}

// List lists the current endpoints and returns a list of handles
func (e *Endpoint) List(filter *ari.Key) (ex []*ari.Key, err error) {
	endpoints := []struct {
		Tech     string `json:"technology"`
		Resource string `json:"resource"`
	}{}

	if filter == nil {
		filter = ari.NodeKey(e.client.ApplicationName(), e.client.node)
	}

	if err = e.client.get("/endpoints", &endpoints); err != nil {
		return nil, err
	}

	for _, i := range endpoints {
		k := e.client.stamp(ari.NewEndpointKey(i.Tech, i.Resource))
		if filter.Match(k) {
			ex = append(ex, k)
		}
	}

	return
}

// ListByTech lists the current endpoints with the given technology and
// returns a list of handles.
func (e *Endpoint) ListByTech(tech string, filter *ari.Key) (ex []*ari.Key, err error) {
	endpoints := []struct {
		Tech     string `json:"technology"`
		Resource string `json:"resource"`
	}{}

	if filter == nil {
		filter = ari.NodeKey(e.client.ApplicationName(), e.client.node)
	}

	if err = e.client.get("/endpoints/"+tech, &endpoints); err != nil {
		return nil, err
	}

	for _, i := range endpoints {
		k := e.client.stamp(ari.NewEndpointKey(i.Tech, i.Resource))
		if filter.Match(k) {
			ex = append(ex, k)
		}
	}

	return
}

// Data retrieves the current state of the endpoint
func (e *Endpoint) Data(key *ari.Key) (*ari.EndpointData, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("endpoint key not supplied")
	}

	if key.Kind != ari.EndpointKey {
		return nil, errors.New("wrong key type")
	}

	data := new(ari.EndpointData)
	if err := e.client.get("/endpoints/"+key.ID, data); err != nil {
		return nil, dataGetError(err, "endpoint", "%s", key.ID)
	}

	data.Key = e.client.stamp(key)

	return data, nil
}
