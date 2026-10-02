// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/rid"
)

// GetVariable reads a bridge variable or function.
func (b *Bridge) GetVariable(key *ari.Key, name string) (string, error) {
	if key == nil || key.ID == "" {
		return "", errors.New("bridge key not supplied")
	}
	query := url.Values{"variable": {name}}
	var response struct {
		Value string `json:"value"`
	}
	path := "/bridges/" + url.PathEscape(key.ID) + "/variable?" + query.Encode()
	if err := b.client.get(path, &response); err != nil {
		return "", err
	}
	return response.Value, nil
}

// SetVariable writes a bridge variable, optionally reporting later changes in events.
func (b *Bridge) SetVariable(key *ari.Key, name, value string, reportEvents *bool) error {
	if key == nil || key.ID == "" {
		return errors.New("bridge key not supplied")
	}
	query := url.Values{"variable": {name}, "value": {value}}
	if reportEvents != nil {
		query.Set("report_events", strconv.FormatBool(*reportEvents))
	}
	path := "/bridges/" + url.PathEscape(key.ID) + "/variable?" + query.Encode()
	return b.client.post(path, nil, nil)
}

// GetVariables reads multiple bridge variables using repeated query parameters.
func (b *Bridge) GetVariables(key *ari.Key, names ...string) (map[string]any, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("bridge key not supplied")
	}
	query := url.Values{"variables": names}
	var response struct {
		Variables map[string]any `json:"variables"`
	}
	path := "/bridges/" + url.PathEscape(key.ID) + "/variables?" + query.Encode()
	if err := b.client.get(path, &response); err != nil {
		return nil, err
	}
	return response.Variables, nil
}

// SetVariables writes a dictionary of bridge variables in one request.
func (b *Bridge) SetVariables(key *ari.Key, values map[string]ari.BridgeVariableAssignment) error {
	if key == nil || key.ID == "" {
		return errors.New("bridge key not supplied")
	}
	request := struct {
		Variables map[string]ari.BridgeVariableAssignment `json:"variables"`
	}{Variables: values}
	return b.client.post("/bridges/"+url.PathEscape(key.ID)+"/variables", nil, request)
}

// Bridge provides the ARI Bridge accessors for the native client
type Bridge struct {
	client *Client
}

// Create creates a bridge and returns the lazy handle for the bridge
func (b *Bridge) Create(key *ari.Key, t string, name string) (bh *ari.BridgeHandle, err error) {
	return b.CreateWithOptions(key, ari.BridgeCreateOptions{Type: t, Name: name})
}

func (b *Bridge) CreateWithOptions(key *ari.Key, opts ari.BridgeCreateOptions) (bh *ari.BridgeHandle, err error) {
	bh, err = b.StageCreateWithOptions(key, opts)
	if err != nil {
		return nil, err
	}

	return bh, bh.Exec()
}

func (b *Bridge) CreateWithoutID(reference *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	return b.CreateOnCollection(reference, "", opts)
}

func (b *Bridge) CreateOnCollection(reference *ari.Key, bridgeID string, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	request := struct {
		Type     string `json:"type,omitempty"`
		Name     string `json:"name,omitempty"`
		BridgeID string `json:"bridgeId,omitempty"`
	}{Type: opts.Type, Name: opts.Name, BridgeID: bridgeID}
	var response struct {
		ID string `json:"id"`
	}
	path := "/bridges"
	var body any = &request
	if len(opts.Variables) > 0 {
		// Asterisk's collection create handler passes the entire JSON body to
		// its variable parser. A nested "variables" object is rejected with 400
		// after the bridge is created. Keep the path-ID route's nested body.
		query := url.Values{}
		if opts.Type != "" {
			query.Set("type", opts.Type)
		}
		if opts.Name != "" {
			query.Set("name", opts.Name)
		}
		if bridgeID != "" {
			query.Set("bridgeId", bridgeID)
		}
		path += "?" + query.Encode()
		body = opts.Variables
	}
	if err := b.client.post(path, &response, body); err != nil {
		return nil, err
	}
	if response.ID == "" {
		return nil, errors.New("bridge creation response omitted id")
	}
	key := ari.NewKey(ari.BridgeKey, response.ID)
	if reference != nil {
		key = reference.New(ari.BridgeKey, response.ID)
	}
	return ari.NewBridgeHandle(b.client.stamp(key), b, nil), nil
}

// StageCreate creates a new bridge handle, staged with a bridge `Create` operation.
func (b *Bridge) StageCreate(key *ari.Key, btype, name string) (*ari.BridgeHandle, error) {
	return b.StageCreateWithOptions(key, ari.BridgeCreateOptions{Type: btype, Name: name})
}

func (b *Bridge) StageCreateWithOptions(key *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	if key == nil {
		key = ari.NewKey(ari.BridgeKey, "")
	}
	if key.ID == "" {
		key.ID = rid.New(rid.Bridge)
	}

	req := struct {
		Type      string                              `json:"type,omitempty"`
		Name      string                              `json:"name,omitempty"`
		Variables map[string]ari.BridgeCreateVariable `json:"variables,omitempty"`
	}{
		Type:      opts.Type,
		Name:      opts.Name,
		Variables: opts.Variables,
	}
	path := "/bridges/" + url.PathEscape(key.ID)

	return ari.NewBridgeHandle(b.client.stamp(key), b, func(bh *ari.BridgeHandle) (err error) {
		return b.client.post(path, nil, &req)
	}), nil
}

// Get gets the lazy handle for the given bridge id
func (b *Bridge) Get(key *ari.Key) *ari.BridgeHandle {
	return ari.NewBridgeHandle(b.client.stamp(key), b, nil)
}

// List lists the current bridges and returns a list of lazy handles
func (b *Bridge) List(filter *ari.Key) (bx []*ari.Key, err error) {
	// native client ignores filter
	bridges := []struct {
		ID string `json:"id"`
	}{}

	err = b.client.get("/bridges", &bridges)

	for _, i := range bridges {
		k := b.client.stamp(ari.NewKey(ari.BridgeKey, i.ID))
		if filter.Match(k) {
			bx = append(bx, k)
		}
	}

	return
}

// Data returns the details of a bridge
// Equivalent to Get /bridges/{bridgeId}
func (b *Bridge) Data(key *ari.Key) (*ari.BridgeData, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("bridge key not supplied")
	}

	data := new(ari.BridgeData)
	if err := b.client.get("/bridges/"+key.ID, data); err != nil {
		return nil, dataGetError(err, "bridge", "%v", key.ID)
	}

	data.Key = b.client.stamp(key)

	return data, nil
}

// AddChannel adds a channel to a bridge
// Equivalent to Post /bridges/{id}/addChannel
func (b *Bridge) AddChannel(key *ari.Key, channelID string) (err error) {
	return b.AddChannelWithOptions(key, channelID, nil)
}

// AddChannelWithOptions adds a channel to a bridge, specifying additional options to be applied to that channel
func (b *Bridge) AddChannelWithOptions(key *ari.Key, channelID string, options *ari.BridgeAddChannelOptions) error {
	if options == nil {
		options = new(ari.BridgeAddChannelOptions)
	}

	req := struct {
		AbsorbDTMF                  bool   `json:"absorbDTMF,omitempty"`
		ChannelID                   string `json:"channel"`
		Mute                        bool   `json:"mute,omitempty"`
		Role                        string `json:"role,omitempty"`
		InhibitConnectedLineUpdates *bool  `json:"inhibitConnectedLineUpdates,omitempty"`
	}{
		AbsorbDTMF:                  options.AbsorbDTMF,
		ChannelID:                   channelID,
		Mute:                        options.Mute,
		Role:                        options.Role,
		InhibitConnectedLineUpdates: options.InhibitConnectedLineUpdates,
	}

	path := "/bridges/" + url.PathEscape(key.ID) + "/addChannel"
	return b.client.post(path, nil, &req)
}

// RemoveChannel removes the specified channel from a bridge
// Equivalent to Post /bridges/{id}/removeChannel
func (b *Bridge) RemoveChannel(key *ari.Key, channelID string) (err error) {
	if key == nil || key.ID == "" {
		return errors.New("bridge key not supplied")
	}
	path := "/bridges/" + url.PathEscape(key.ID) + "/removeChannel"
	return b.client.post(path, nil, &struct {
		Channel string `json:"channel"`
	}{Channel: channelID})
}

// Delete shuts down a bridge. If any channels are in this bridge,
// they will be removed and resume whatever they were doing beforehand.
// This means that the channels themselves are not deleted.
// Equivalent to DELETE /bridges/{id}
func (b *Bridge) Delete(key *ari.Key) (err error) {
	return b.client.del("/bridges/"+key.ID, nil, "")
}

// MOH requests that the given musiconhold class be played to the bridge
func (b *Bridge) MOH(key *ari.Key, class string) error {
	if key == nil || key.ID == "" {
		return errors.New("bridge key not supplied")
	}
	path := "/bridges/" + url.PathEscape(key.ID) + "/moh"
	return b.client.post(path, nil, &struct {
		Class string `json:"mohClass"`
	}{Class: class})
}

// StopMOH requests that any MusicOnHold which is playing to the bridge be stopped.
func (b *Bridge) StopMOH(key *ari.Key) error {
	return b.client.del("/bridges/"+key.ID+"/moh", nil, "")
}

// Play attempts to play the given mediaURI on the bridge, using the playbackID
// as the identifier to the created playback handle
func (b *Bridge) Play(key *ari.Key, playbackID string, mediaURI ...string) (*ari.PlaybackHandle, error) {
	return b.PlayWithOptions(key, playbackID, ari.BridgePlayOptions{Media: mediaURI})
}

func (b *Bridge) PlayWithOptions(key *ari.Key, playbackID string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	if playbackID == "" {
		playbackID = rid.New(rid.Playback)
	}

	h, err := b.StagePlayWithOptions(key, playbackID, opts)
	if err != nil {
		return nil, err
	}

	return h, h.Exec()
}

// StagePlay stages a `Play` operation on the bridge
func (b *Bridge) StagePlay(key *ari.Key, playbackID string, mediaURI ...string) (*ari.PlaybackHandle, error) {
	return b.StagePlayWithOptions(key, playbackID, ari.BridgePlayOptions{Media: mediaURI})
}

func (b *Bridge) StagePlayWithOptions(key *ari.Key, playbackID string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	if playbackID == "" {
		playbackID = rid.New(rid.Playback)
	}

	resp := make(map[string]interface{})
	playbackKey := b.client.stamp(ari.NewKey(ari.PlaybackKey, playbackID))
	path := "/bridges/" + url.PathEscape(key.ID) + "/play/" + url.PathEscape(playbackID)
	request := bridgePlayBody(opts)

	return ari.NewPlaybackHandle(playbackKey, b.client.Playback(), func(h *ari.PlaybackHandle) error {
		return b.client.post(path, &resp, &request)
	}), nil
}

type bridgePlayRequest struct {
	Media           []string `json:"media"`
	AnnouncerFormat string   `json:"announcer_format,omitempty"`
	Lang            string   `json:"lang,omitempty"`
	OffsetMS        *int     `json:"offsetms,omitempty"`
	SkipMS          *int     `json:"skipms,omitempty"`
	PlaybackID      string   `json:"playbackId,omitempty"`
}

func bridgePlayBody(opts ari.BridgePlayOptions) bridgePlayRequest {
	return bridgePlayRequest{Media: opts.Media, AnnouncerFormat: opts.AnnouncerFormat, Lang: opts.Lang, OffsetMS: opts.OffsetMS, SkipMS: opts.SkipMS}
}

func (b *Bridge) PlayWithoutID(key *ari.Key, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	return b.PlayOnCollection(key, "", opts)
}

func (b *Bridge) PlayOnCollection(key *ari.Key, playbackID string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("bridge key not supplied")
	}
	path := "/bridges/" + url.PathEscape(key.ID) + "/play"
	request := bridgePlayBody(opts)
	request.PlaybackID = playbackID
	var response struct {
		ID string `json:"id"`
	}
	if err := b.client.post(path, &response, &request); err != nil {
		return nil, err
	}
	if response.ID == "" {
		return nil, errors.New("bridge playback response omitted id")
	}
	return ari.NewPlaybackHandle(b.client.stamp(key.New(ari.PlaybackKey, response.ID)), b.client.Playback(), nil), nil
}

// Record attempts to record audio on the bridge, using name as the identifier for
// the created live recording handle
func (b *Bridge) Record(key *ari.Key, name string, opts *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	h, err := b.StageRecord(key, name, opts)
	if err != nil {
		return nil, err
	}

	return h, h.Exec()
}

// StageRecord stages a `Record` opreation
func (b *Bridge) StageRecord(key *ari.Key, name string, opts *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("bridge key not supplied")
	}
	if opts == nil {
		opts = &ari.RecordingOptions{}
	}

	resp := make(map[string]interface{})
	request := struct {
		Name           string `json:"name"`
		Format         string `json:"format"`
		RecorderFormat string `json:"recorder_format,omitempty"`
		MaxDuration    int    `json:"maxDurationSeconds"`
		MaxSilence     int    `json:"maxSilenceSeconds"`
		IfExists       string `json:"ifExists,omitempty"`
		Beep           bool   `json:"beep"`
		TerminateOn    string `json:"terminateOn,omitempty"`
	}{name, opts.Format, opts.RecorderFormat, int(opts.MaxDuration / time.Second), int(opts.MaxSilence / time.Second), opts.Exists, opts.Beep, opts.Terminate}

	recordingKey := b.client.stamp(ari.NewKey(ari.LiveRecordingKey, name))

	return ari.NewLiveRecordingHandle(recordingKey, b.client.LiveRecording(), func(h *ari.LiveRecordingHandle) error {
		path := "/bridges/" + url.PathEscape(key.ID) + "/record"
		return b.client.post(path, &resp, &request)
	}), nil
}

// Subscribe creates an event subscription for events related to the given
// bridge␃entity
func (b *Bridge) Subscribe(key *ari.Key, n ...string) ari.Subscription {
	return b.client.Bus().Subscribe(key, n...)
}

// VideoSource sets a channel as the video source in a multi-party mixing bridge.
// This operation has no effect on bridges with two or fewer participants.
// Equivalent to POST /bridges/{bridgeId}/videoSource/{channelId}
func (b *Bridge) VideoSource(key *ari.Key, channelID string) error {
	return b.client.post("/bridges/"+key.ID+"/videoSource/"+channelID, nil, nil)
}

// VideoSourceDelete removes any explicit video source in a multi-party mixing bridge.
// This operation has no effect on bridges with two or fewer participants.
// When no explicit video source is set, talk detection will be used to determine the active video stream.
// Equivalent to DELETE /bridges/{bridgeId}/videoSource
func (b *Bridge) VideoSourceDelete(key *ari.Key) error {
	return b.client.del("/bridges/"+key.ID+"/videoSource", nil, "")
}
