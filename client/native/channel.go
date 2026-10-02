package native

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/rid"
)

// Channel provides the ARI Channel accessors for the native client
type Channel struct {
	client *Client
}

// List lists the current channels and returns the list of channel handles
func (c *Channel) List(filter *ari.Key) (cx []*ari.Key, err error) {
	channels := []struct {
		ID string `json:"id"`
	}{}

	if filter == nil {
		filter = ari.NewKey(ari.ChannelKey, "")
	}

	err = c.client.get("/channels", &channels)

	for _, i := range channels {
		k := c.client.stamp(ari.NewKey(ari.ChannelKey, i.ID))
		if filter.Match(k) {
			cx = append(cx, k)
		}
	}

	return
}

// Hangup hangs up the given channel using the (optional) reason.
func (c *Channel) Hangup(key *ari.Key, reason string) error {
	if reason == "" {
		reason = "normal"
	}
	return c.HangupWithOptions(key, ari.ChannelHangupOptions{Reason: reason})
}

func (c *Channel) HangupWithOptions(key *ari.Key, opts ari.ChannelHangupOptions) error {
	if key == nil || key.ID == "" {
		return errors.New("channel key not supplied")
	}
	query := url.Values{}
	if opts.Reason != "" {
		query.Set("reason", opts.Reason)
	}
	if opts.ReasonCode != "" {
		query.Set("reason_code", opts.ReasonCode)
	}
	return c.client.del("/channels/"+url.PathEscape(key.ID), nil, query.Encode())
}

// Data retrieves the current state of the channel
func (c *Channel) Data(key *ari.Key) (*ari.ChannelData, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("channel key not supplied")
	}

	data := new(ari.ChannelData)
	if err := c.client.get("/channels/"+key.ID, data); err != nil {
		return nil, dataGetError(err, "channel", "%v", key.ID)
	}

	data.Key = c.client.stamp(key)

	return data, nil
}

// Get gets the lazy handle for the given channel
func (c *Channel) Get(key *ari.Key) *ari.ChannelHandle {
	return ari.NewChannelHandle(c.client.stamp(key), c, nil)
}

// Originate originates a channel and returns the handle.
//
// **Note** that referenceKey is completely optional.  It is used for placing
// the new channel onto the correct Asterisk node and for assigning default
// values for communications parameters such as codecs.
func (c *Channel) Originate(referenceKey *ari.Key, req ari.OriginateRequest) (*ari.ChannelHandle, error) {
	h, err := c.StageOriginate(referenceKey, req)
	if err != nil {
		return nil, err
	}

	return h, h.Exec()
}

func (c *Channel) OriginateWithID(referenceKey *ari.Key, req ari.OriginateRequest) (*ari.ChannelHandle, error) {
	if req.ChannelID == "" {
		return nil, errors.New("channel ID required for path-ID originate")
	}
	if referenceKey != nil && req.Originator == "" && referenceKey.Kind == ari.ChannelKey {
		req.Originator = referenceKey.ID
	}
	path := "/channels/" + url.PathEscape(req.ChannelID)
	var response struct {
		ID string `json:"id"`
	}
	if err := c.client.post(path, &response, &req); err != nil {
		return nil, err
	}
	if response.ID != "" && response.ID != req.ChannelID {
		return nil, fmt.Errorf("originated channel ID %q differs from requested %q", response.ID, req.ChannelID)
	}
	key := ari.NewKey(ari.ChannelKey, req.ChannelID)
	if referenceKey != nil {
		key = referenceKey.New(ari.ChannelKey, req.ChannelID)
	}
	return ari.NewChannelHandle(c.client.stamp(key), c, nil), nil
}

// StageOriginate creates a new channel handle with a channel originate request
// staged.
//
// **Note** that referenceKey is completely optional.  It is used for placing
// the new channel onto the correct Asterisk node and for assigning default
// values for communications parameters such as codecs.
func (c *Channel) StageOriginate(referenceKey *ari.Key, req ari.OriginateRequest) (*ari.ChannelHandle, error) {
	if referenceKey != nil && req.Originator == "" && referenceKey.Kind == ari.ChannelKey {
		req.Originator = referenceKey.ID
	}

	if req.ChannelID == "" {
		req.ChannelID = rid.New(rid.Channel)
	}
	return ari.NewChannelHandle(c.client.stamp(ari.NewKey(ari.ChannelKey, req.ChannelID)), c,
		func(ch *ari.ChannelHandle) error {
			type response struct {
				ID string `json:"id"`
			}

			var resp response

			return c.client.post("/channels", &resp, &req)
		},
	), nil
}

// Create creates a channel and returns the handle. TODO: expand
// differences between originate and create.
func (c *Channel) Create(key *ari.Key, req ari.ChannelCreateRequest) (*ari.ChannelHandle, error) {
	if key != nil && req.Originator == "" && key.Kind == ari.ChannelKey {
		req.Originator = key.ID
	}

	if req.ChannelID == "" {
		req.ChannelID = rid.New(rid.Channel)
	}

	err := c.client.post("/channels/create", nil, &req)
	if err != nil {
		return nil, err
	}

	return ari.NewChannelHandle(c.client.stamp(ari.NewKey(ari.ChannelKey, req.ChannelID)), c, nil), nil
}

// Continue tells a channel to process to the given ARI context and extension
func (c *Channel) Continue(key *ari.Key, context, extension string, priority int) (err error) {
	return c.ContinueWithOptions(key, ari.ChannelContinueOptions{Context: context, Extension: extension, Priority: &priority})
}

func (c *Channel) ContinueWithOptions(key *ari.Key, opts ari.ChannelContinueOptions) error {
	if key == nil || key.ID == "" {
		return errors.New("channel key not supplied")
	}
	path := "/channels/" + url.PathEscape(key.ID) + "/continue"
	return c.client.post(path, nil, &struct {
		Context   string `json:"context,omitempty"`
		Extension string `json:"extension,omitempty"`
		Priority  *int   `json:"priority,omitempty"`
		Label     string `json:"label,omitempty"`
	}{opts.Context, opts.Extension, opts.Priority, opts.Label})
}

// Move moves the channel to another stasis application
func (c *Channel) Move(key *ari.Key, app string, appArgs string) error {
	if key == nil || key.ID == "" {
		return errors.New("channel key not supplied")
	}
	return c.client.post("/channels/"+url.PathEscape(key.ID)+"/move", nil, &struct {
		App     string `json:"app"`
		AppArgs string `json:"appArgs,omitempty"`
	}{app, appArgs})
}

// Busy sends the busy status code to the channel (TODO: does this play a busy signal too)
func (c *Channel) Busy(key *ari.Key) error {
	return c.Hangup(key, "busy")
}

// Congestion sends the congestion status code to the channel (TODO: does this play a tone?)
func (c *Channel) Congestion(key *ari.Key) error {
	return c.Hangup(key, "congestion")
}

// Answer answers a channel, if ringing (TODO: does this return an error if already answered?)
func (c *Channel) Answer(key *ari.Key) error {
	return c.client.post("/channels/"+key.ID+"/answer", nil, nil)
}

// Ring causes a channel to start ringing (TODO: does this return an error if already ringing?)
func (c *Channel) Ring(key *ari.Key) error {
	return c.client.post("/channels/"+key.ID+"/ring", nil, nil)
}

// StopRing causes a channel to stop ringing (TODO: does this return an error if not ringing?)
func (c *Channel) StopRing(key *ari.Key) error {
	return c.client.del("/channels/"+key.ID+"/ring", nil, "")
}

// Hold puts a channel on hold (TODO: does this return an error if already on hold?)
func (c *Channel) Hold(key *ari.Key) error {
	return c.client.post("/channels/"+key.ID+"/hold", nil, nil)
}

// StopHold removes a channel from hold (TODO: does this return an error if not on hold)
func (c *Channel) StopHold(key *ari.Key) (err error) {
	return c.client.del("/channels/"+key.ID+"/hold", nil, "")
}

// Mute mutes a channel in the given direction (TODO: does this return an error if already muted)
func (c *Channel) Mute(key *ari.Key, dir ari.Direction) error {
	if key == nil || key.ID == "" {
		return errors.New("channel key not supplied")
	}
	if dir == "" {
		dir = ari.DirectionBoth
	}
	return c.client.post("/channels/"+url.PathEscape(key.ID)+"/mute", nil, &struct {
		Direction ari.Direction `json:"direction"`
	}{dir})
}

// Unmute unmutes a channel in the given direction (TODO: does this return an error if unmuted)
func (c *Channel) Unmute(key *ari.Key, dir ari.Direction) (err error) {
	if key == nil || key.ID == "" {
		return errors.New("channel key not supplied")
	}
	if dir == "" {
		dir = ari.DirectionBoth
	}
	return c.client.del("/channels/"+url.PathEscape(key.ID)+"/mute", nil, url.Values{"direction": {string(dir)}}.Encode())
}

// SendDTMF sends a string of digits and symbols to the channel
func (c *Channel) SendDTMF(key *ari.Key, dtmf string, opts *ari.DTMFOptions) error {
	options := ari.DTMFOptions{}
	if opts != nil {
		options = *opts
	}

	if options.Duration <= 0 {
		options.Duration = 100 * time.Millisecond
	}

	if options.Between <= 0 {
		options.Between = 100 * time.Millisecond
	}

	req := struct {
		Dtmf     string `json:"dtmf,omitempty"`
		Before   int    `json:"before,omitempty"`
		Between  int    `json:"between,omitempty"`
		Duration int    `json:"duration,omitempty"`
		After    int    `json:"after,omitempty"`
	}{
		Dtmf:     dtmf,
		Before:   int(options.Before / time.Millisecond),
		After:    int(options.After / time.Millisecond),
		Duration: int(options.Duration / time.Millisecond),
		Between:  int(options.Between / time.Millisecond),
	}

	return c.client.post("/channels/"+key.ID+"/dtmf", nil, &req)
}

// MOH plays the given music on hold class to the channel TODO: does this error when already playing MOH?
func (c *Channel) MOH(key *ari.Key, class string) error {
	if key == nil || key.ID == "" {
		return errors.New("channel key not supplied")
	}
	path := "/channels/" + url.PathEscape(key.ID) + "/moh"
	return c.client.post(path, nil, &struct {
		Class string `json:"mohClass,omitempty"`
	}{class})
}

// StopMOH stops any music on hold playing on the channel (TODO: does this error when no MOH is playing?)
func (c *Channel) StopMOH(key *ari.Key) error {
	return c.client.del("/channels/"+key.ID+"/moh", nil, "")
}

// Silence silences a channel (TODO: does this error when already silences)
func (c *Channel) Silence(key *ari.Key) error {
	return c.client.post("/channels/"+key.ID+"/silence", nil, nil)
}

// StopSilence stops the silence on a channel (TODO: does this error when not silenced)
func (c *Channel) StopSilence(key *ari.Key) error {
	return c.client.del("/channels/"+key.ID+"/silence", nil, "")
}

// Play plays the given media URI on the channel, using the playbackID as
// the identifier of the created ARI Playback entity
func (c *Channel) Play(key *ari.Key, playbackID string, mediaURI ...string) (*ari.PlaybackHandle, error) {
	return c.PlayWithOptions(key, playbackID, ari.ChannelPlayOptions{Media: mediaURI})
}

func (c *Channel) PlayWithOptions(key *ari.Key, playbackID string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	if playbackID == "" {
		playbackID = rid.New(rid.Playback)
	}

	h, err := c.StagePlayWithOptions(key, playbackID, opts)
	if err != nil {
		return nil, err
	}

	return h, h.Exec()
}

// StagePlay stages a `Play` operation on the bridge
func (c *Channel) StagePlay(key *ari.Key, playbackID string, mediaURI ...string) (*ari.PlaybackHandle, error) {
	return c.StagePlayWithOptions(key, playbackID, ari.ChannelPlayOptions{Media: mediaURI})
}

func (c *Channel) StagePlayWithOptions(key *ari.Key, playbackID string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	if playbackID == "" {
		playbackID = rid.New(rid.Playback)
	}

	resp := make(map[string]interface{})

	playbackKey := c.client.stamp(ari.NewKey(ari.PlaybackKey, playbackID))
	path := "/channels/" + url.PathEscape(key.ID) + "/play/" + url.PathEscape(playbackID)
	request := channelPlayBody(opts)

	return ari.NewPlaybackHandle(playbackKey, c.client.Playback(), func(pb *ari.PlaybackHandle) error {
		return c.client.post(path, &resp, &request)
	}), nil
}

type channelPlayRequest struct {
	Media      []string `json:"media"`
	Lang       string   `json:"lang,omitempty"`
	OffsetMS   *int     `json:"offsetms,omitempty"`
	SkipMS     *int     `json:"skipms,omitempty"`
	PlaybackID string   `json:"playbackId,omitempty"`
}

func channelPlayBody(opts ari.ChannelPlayOptions) channelPlayRequest {
	return channelPlayRequest{Media: opts.Media, Lang: opts.Lang, OffsetMS: opts.OffsetMS, SkipMS: opts.SkipMS}
}

func (c *Channel) PlayWithoutID(key *ari.Key, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	return c.PlayOnCollection(key, "", opts)
}

func (c *Channel) PlayOnCollection(key *ari.Key, playbackID string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("channel key not supplied")
	}
	path := "/channels/" + url.PathEscape(key.ID) + "/play"
	request := channelPlayBody(opts)
	request.PlaybackID = playbackID
	var response struct {
		ID string `json:"id"`
	}
	if err := c.client.post(path, &response, &request); err != nil {
		return nil, err
	}
	if response.ID == "" {
		return nil, errors.New("channel playback response omitted id")
	}
	return ari.NewPlaybackHandle(c.client.stamp(key.New(ari.PlaybackKey, response.ID)), c.client.Playback(), nil), nil
}

// Record records audio on the channel, using the name parameter as the name of the
// created LiveRecording entity.
func (c *Channel) Record(key *ari.Key, name string, opts *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	h, err := c.StageRecord(key, name, opts)
	if err != nil {
		return nil, err
	}

	return h, h.Exec()
}

// StageRecord stages a `Record` opreation
func (c *Channel) StageRecord(key *ari.Key, name string, opts *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	if opts == nil {
		opts = &ari.RecordingOptions{}
	}

	resp := make(map[string]interface{})
	req := struct {
		Name        string `json:"name"`
		Format      string `json:"format"`
		MaxDuration int    `json:"maxDurationSeconds"`
		MaxSilence  int    `json:"maxSilenceSeconds"`
		IfExists    string `json:"ifExists,omitempty"`
		Beep        bool   `json:"beep"`
		TerminateOn string `json:"terminateOn,omitempty"`
	}{
		Name:        name,
		Format:      opts.Format,
		MaxDuration: int(opts.MaxDuration / time.Second),
		MaxSilence:  int(opts.MaxSilence / time.Second),
		IfExists:    opts.Exists,
		Beep:        opts.Beep,
		TerminateOn: opts.Terminate,
	}

	recordingKey := c.client.stamp(ari.NewKey(ari.LiveRecordingKey, name))

	return ari.NewLiveRecordingHandle(recordingKey, c.client.LiveRecording(), func(h *ari.LiveRecordingHandle) error {
		return c.client.post("/channels/"+key.ID+"/record", &resp, &req)
	}), nil
}

// Snoop snoops on a channel, using the the given snoopID as the new channel handle ID (TODO: confirm and expand description)
func (c *Channel) Snoop(key *ari.Key, snoopID string, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	h, err := c.StageSnoop(key, snoopID, opts)
	if err != nil {
		return nil, err
	}

	return h, h.Exec()
}

func (c *Channel) SnoopWithoutID(key *ari.Key, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	return c.SnoopOnCollection(key, "", opts)
}

func (c *Channel) SnoopOnCollection(key *ari.Key, snoopID string, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("channel key not supplied")
	}
	if opts == nil {
		opts = &ari.SnoopOptions{App: c.client.ApplicationName()}
	}
	var response struct {
		ID string `json:"id"`
	}
	path := "/channels/" + url.PathEscape(key.ID) + "/snoop"
	request := snoopBody(opts, snoopID)
	if err := c.client.post(path, &response, &request); err != nil {
		return nil, err
	}
	if response.ID == "" {
		return nil, errors.New("snoop response omitted id")
	}
	return ari.NewChannelHandle(c.client.stamp(key.New(ari.ChannelKey, response.ID)), c, nil), nil
}

// StageSnoop creates a new `ChannelHandle` with a `Snoop` operation staged.
func (c *Channel) StageSnoop(key *ari.Key, snoopID string, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("channel key not supplied")
	}
	if opts == nil {
		opts = &ari.SnoopOptions{App: c.client.ApplicationName()}
	}

	if snoopID == "" {
		snoopID = rid.New(rid.Snoop)
	}

	// Create the snooping channel's key
	k := c.client.stamp(ari.NewKey(ari.ChannelKey, snoopID))
	path := "/channels/" + url.PathEscape(key.ID) + "/snoop/" + url.PathEscape(snoopID)
	request := snoopBody(opts, "")

	return ari.NewChannelHandle(k, c, func(ch *ari.ChannelHandle) error {
		return c.client.post(path, nil, &request)
	}), nil
}

type snoopRequest struct {
	App     string        `json:"app"`
	AppArgs string        `json:"appArgs,omitempty"`
	Spy     ari.Direction `json:"spy,omitempty"`
	Whisper ari.Direction `json:"whisper,omitempty"`
	SnoopID string        `json:"snoopId,omitempty"`
}

func snoopBody(opts *ari.SnoopOptions, snoopID string) snoopRequest {
	return snoopRequest{App: opts.App, AppArgs: opts.AppArgs, Spy: opts.Spy, Whisper: opts.Whisper, SnoopID: snoopID}
}

// ExternalMedia implements the ari.Channel interface
func (c *Channel) ExternalMedia(key *ari.Key, opts ari.ExternalMediaOptions) (*ari.ChannelHandle, error) {
	h, err := c.StageExternalMedia(key, opts)
	if err != nil {
		return nil, err
	}

	return h, h.Exec()
}

// StageExternalMedia implements the ari.Channel interface
func (c *Channel) StageExternalMedia(key *ari.Key, opts ari.ExternalMediaOptions) (*ari.ChannelHandle, error) {
	if opts.ChannelID == "" {
		opts.ChannelID = rid.New(rid.Channel)
	}

	if opts.App == "" {
		opts.App = c.client.ApplicationName()
	}

	if opts.ExternalHost == "" {
		return nil, errors.New("ExternalHost is mandatory")
	}

	if opts.Encapsulation == "" {
		opts.Encapsulation = "rtp"
	}

	if opts.Transport == "" {
		opts.Transport = "udp"
	}

	if opts.ConnectionType == "" {
		opts.ConnectionType = "client"
	}

	if opts.Format == "" {
		return nil, errors.New("format is mandatory")
	}

	if opts.Direction == "" {
		opts.Direction = "both"
	}

	// Create the snooping channel's key
	k := c.client.stamp(ari.NewKey(ari.ChannelKey, opts.ChannelID))

	return ari.NewChannelHandle(k, c, func(ch *ari.ChannelHandle) error {
		return c.client.post("/channels/externalMedia", nil, &opts)
	}), nil
}

// Dial dials the given calling channel identifier
func (c *Channel) Dial(key *ari.Key, callingChannelID string, timeout time.Duration) error {
	if key == nil || key.ID == "" {
		return errors.New("channel key not supplied")
	}
	return c.client.post("/channels/"+url.PathEscape(key.ID)+"/dial", nil, &struct {
		Caller  string `json:"caller,omitempty"`
		Timeout int    `json:"timeout"`
	}{callingChannelID, int(timeout / time.Second)})
}

// Subscribe creates a new subscription for ARI events related to this channel
func (c *Channel) Subscribe(key *ari.Key, n ...string) ari.Subscription {
	return c.client.Bus().Subscribe(key, n...)
}

// GetVariable gets the value of the given variable
func (c *Channel) GetVariable(key *ari.Key, name string) (string, error) {
	var m struct {
		Value string `json:"value"`
	}

	query := url.Values{"variable": {name}}
	err := c.client.get(fmt.Sprintf("/channels/%s/variable?%s", key.ID, query.Encode()), &m)

	return m.Value, err
}

// SetVariable sets the value of the given channel variable
func (c *Channel) SetVariable(key *ari.Key, name, value string) error {
	return c.SetVariableWithOptions(key, name, value, nil)
}

func (c *Channel) SetVariableWithOptions(key *ari.Key, name, value string, opts *ari.ChannelVariableSetOptions) error {
	if key == nil || key.ID == "" {
		return errors.New("channel key not supplied")
	}
	if name == "" {
		return errors.New("variable name not supplied")
	}
	var reportEvents *bool
	if opts != nil {
		reportEvents = opts.ReportEvents
	}
	return c.client.post("/channels/"+url.PathEscape(key.ID)+"/variable", nil, &struct {
		Variable     string `json:"variable"`
		Value        string `json:"value"`
		ReportEvents *bool  `json:"report_events,omitempty"`
	}{name, value, reportEvents})
}

func channelPath(key *ari.Key) (string, error) {
	if key == nil || key.ID == "" {
		return "", errors.New("channel key not supplied")
	}
	return "/channels/" + url.PathEscape(key.ID), nil
}

// GetVariables retrieves several channel variables in one request.
func (c *Channel) GetVariables(key *ari.Key, names ...string) (map[string]any, error) {
	path, err := channelPath(key)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, errors.New("variable names not supplied")
	}
	query := url.Values{"variables": names}
	var result struct {
		Variables map[string]any `json:"variables"`
	}
	err = c.client.get(path+"/variables?"+query.Encode(), &result)
	return result.Variables, err
}

// SetVariables assigns several channel variables in one request.
func (c *Channel) SetVariables(key *ari.Key, values map[string]ari.VariableAssignment) error {
	path, err := channelPath(key)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return errors.New("variable values not supplied")
	}
	return c.client.post(path+"/variables", nil, &struct {
		Variables map[string]ari.VariableAssignment `json:"variables"`
	}{Variables: values})
}

func (c *Channel) Redirect(key *ari.Key, endpoint string) error {
	path, err := channelPath(key)
	if err != nil {
		return err
	}
	if endpoint == "" {
		return errors.New("endpoint not supplied")
	}
	return c.client.post(path+"/redirect?"+url.Values{"endpoint": {endpoint}}.Encode(), nil, nil)
}

func (c *Channel) Progress(key *ari.Key) error {
	path, err := channelPath(key)
	if err != nil {
		return err
	}
	return c.client.post(path+"/progress", nil, nil)
}

func (c *Channel) TransferProgress(key *ari.Key, state string) error {
	path, err := channelPath(key)
	if err != nil {
		return err
	}
	if state == "" {
		return errors.New("transfer state not supplied")
	}
	return c.client.post(path+"/transfer_progress?"+url.Values{"states": {state}}.Encode(), nil, nil)
}

func (c *Channel) RTPStatistics(key *ari.Key) (*ari.RTPStats, error) {
	path, err := channelPath(key)
	if err != nil {
		return nil, err
	}
	stats := new(ari.RTPStats)
	if err := c.client.get(path+"/rtp_statistics", stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// UserEvent - triggers a UserEvent for the given channel
func (c *Channel) UserEvent(key *ari.Key, ue *ari.ChannelUserevent) error {
	if key == nil || key.ID == "" || key.App == "" {
		return errors.New("channel key with application not supplied")
	}
	if ue == nil || ue.Eventname == "" {
		return errors.New("user event name not supplied")
	}
	body := struct {
		Application string `json:"application"`
		Source      string `json:"source"`
		Variables   any    `json:"variables,omitempty"`
	}{Application: key.App, Source: "channel:" + key.ID, Variables: ue.Userevent}
	return c.client.post("/events/user/"+url.PathEscape(ue.Eventname), nil, &body)
}
