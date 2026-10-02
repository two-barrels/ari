package native

import (
	"net/url"

	"github.com/rotisserie/eris"

	"github.com/two-barrels/ari/v6"
)

// Asterisk provides the ARI Asterisk accessors for a native client
type Asterisk struct {
	client *Client
}

// Logging provides the ARI Asterisk Logging accessors for a native client
func (a *Asterisk) Logging() ari.Logging {
	return &Logging{a.client}
}

// Modules provides the ARI Asterisk Modules accessors for a native client
func (a *Asterisk) Modules() ari.Modules {
	return &Modules{a.client}
}

// Config provides the ARI Asterisk Config accessors for a native client
func (a *Asterisk) Config() ari.Config {
	return &Config{a.client}
}

/*
	conn    *Conn
	logging ari.Logging
	modules ari.Modules
	config  ari.Config
}
*/

// Info returns various data about the Asterisk system
// Equivalent to GET /asterisk/info
func (a *Asterisk) Info(key *ari.Key) (*ari.AsteriskInfo, error) {
	return a.InfoWithOptions(key, ari.AsteriskInfoOptions{})
}

func (a *Asterisk) InfoWithOptions(key *ari.Key, opts ari.AsteriskInfoOptions) (*ari.AsteriskInfo, error) {
	var m ari.AsteriskInfo
	path := "/asterisk/info"
	if opts.Only != "" {
		path += "?" + url.Values{"only": {opts.Only}}.Encode()
	}

	return &m, eris.Wrap(
		a.client.get(path, &m),
		"failed to get asterisk info",
	)
}

// Ping returns the ARI ping response from Asterisk.
func (a *Asterisk) Ping(key *ari.Key) (*ari.AsteriskPing, error) {
	var result ari.AsteriskPing
	if err := a.client.get("/asterisk/ping", &result); err != nil {
		return nil, eris.Wrap(err, "failed to ping asterisk")
	}
	return &result, nil
}

// AsteriskVariables provides the ARI Variables accessors for server-level variables
type AsteriskVariables struct {
	client *Client
}

// Variables returns the variables interface for the Asterisk server
func (a *Asterisk) Variables() ari.AsteriskVariables {
	return &AsteriskVariables{a.client}
}

// Get returns the value of the given global variable
// Equivalent to GET /asterisk/variable
func (a *AsteriskVariables) Get(key *ari.Key) (string, error) {
	var m struct {
		Value string `json:"value"`
	}

	query := url.Values{"variable": {key.ID}}
	err := a.client.get("/asterisk/variable?"+query.Encode(), &m)
	if err != nil {
		return "", eris.Wrapf(err, "Error getting asterisk variable '%v'", key.ID)
	}

	return m.Value, nil
}

// Set sets a global channel variable
// (Equivalent to POST /asterisk/variable)
func (a *AsteriskVariables) Set(key *ari.Key, value string) (err error) {
	if key == nil || key.ID == "" {
		return eris.New("variable key not supplied")
	}
	return eris.Wrapf(
		a.client.post("/asterisk/variable", nil, &struct {
			Variable string `json:"variable"`
			Value    string `json:"value"`
		}{Variable: key.ID, Value: value}),
		"Error setting asterisk variable '%s' to '%s'", key.ID, value,
	)
}
