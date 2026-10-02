package native

import (
	"errors"
	"net/url"

	"github.com/rotisserie/eris"

	"github.com/two-barrels/ari/v6"
)

// ClaimChannel atomically claims a broadcast channel for the application.
func (a *Application) ClaimChannel(key *ari.Key, channelID string) error {
	if key == nil || key.ID == "" {
		return errors.New("application key not supplied")
	}
	if channelID == "" {
		return errors.New("channel ID not supplied")
	}
	query := url.Values{"channelId": {channelID}, "application": {key.ID}}
	return a.client.post("/events/claim?"+query.Encode(), nil, nil)
}

// Application is a native implementation of ARI's Application functions
type Application struct {
	client *Client
}

// Get returns a managed handle to an ARI application
func (a *Application) Get(key *ari.Key) *ari.ApplicationHandle {
	return ari.NewApplicationHandle(a.client.stamp(key), a)
}

// List returns the list of applications managed by asterisk
func (a *Application) List(filter *ari.Key) (ax []*ari.Key, err error) {
	if filter == nil {
		filter = ari.NewKey(ari.ApplicationKey, "")
	}

	apps := []struct {
		Name string `json:"name"`
	}{}

	err = a.client.get("/applications", &apps)

	for _, i := range apps {
		k := a.client.stamp(ari.NewKey(ari.ApplicationKey, i.Name))
		if filter.Match(k) {
			ax = append(ax, k)
		}
	}

	err = eris.Wrap(err, "Error listing applications")

	return
}

// Data returns the details of a given ARI application
// Equivalent to GET /applications/{applicationName}
func (a *Application) Data(key *ari.Key) (*ari.ApplicationData, error) {
	if key == nil || key.ID == "" {
		return nil, eris.New("application key not supplied")
	}

	data := new(ari.ApplicationData)
	if err := a.client.get("/applications/"+key.ID, data); err != nil {
		return nil, dataGetError(err, "application", "%v", key.ID)
	}

	data.Key = a.client.stamp(key)

	return data, nil
}

// Subscribe subscribes the given application to an event source
// Equivalent to POST /applications/{applicationName}/subscription
func (a *Application) Subscribe(key *ari.Key, eventSource string) error {
	if key == nil || key.ID == "" {
		return errors.New("application key not supplied")
	}
	if eventSource == "" {
		return errors.New("event source not supplied")
	}
	err := a.client.post("/applications/"+url.PathEscape(key.ID)+"/subscription", nil, &struct {
		EventSource string `json:"eventSource"`
	}{EventSource: eventSource})

	return eris.Wrapf(err, "Error subscribing application '%v' for event source '%v'", key.ID, eventSource)
}

// Unsubscribe unsubscribes (removes a subscription to) a given
// ARI application from the provided event source
// Equivalent to DELETE /applications/{applicationName}/subscription
func (a *Application) Unsubscribe(key *ari.Key, eventSource string) error {
	name := key.ID
	query := url.Values{"eventSource": {eventSource}}
	err := a.client.del("/applications/"+name+"/subscription", nil, query.Encode())

	return eris.Wrapf(err, "Error unsubscribing application '%v' for event source '%v'", name, eventSource)
}

// FilterEvents updates the application event filter and returns the new state.
func (a *Application) FilterEvents(key *ari.Key, filter *ari.ApplicationEventFilter) (*ari.ApplicationData, error) {
	if key == nil || key.ID == "" {
		return nil, eris.New("application key not supplied")
	}
	if filter == nil {
		filter = &ari.ApplicationEventFilter{}
	}
	var data ari.ApplicationData
	path := "/applications/" + url.PathEscape(key.ID) + "/eventFilter"
	if err := a.client.put(path, &data, filter); err != nil {
		return nil, eris.Wrapf(err, "failed to filter application %q events", key.ID)
	}
	data.Key = a.client.stamp(key)
	return &data, nil
}
