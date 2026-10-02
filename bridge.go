package ari

import "encoding/json"

// Bridge represents a communication path to an
// Asterisk server for working with bridge resources
type Bridge interface {
	// Create creates a bridge
	Create(key *Key, btype string, name string) (*BridgeHandle, error)
	CreateWithOptions(key *Key, opts BridgeCreateOptions) (*BridgeHandle, error)
	// CreateWithoutID uses POST /bridges and returns Asterisk's assigned ID.
	CreateWithoutID(reference *Key, opts BridgeCreateOptions) (*BridgeHandle, error)
	// CreateOnCollection uses POST /bridges with an optional bridgeId query value.
	CreateOnCollection(reference *Key, bridgeID string, opts BridgeCreateOptions) (*BridgeHandle, error)

	// StageCreate creates a new bridge handle, staged with a bridge `Create` operation.
	StageCreate(key *Key, btype string, name string) (*BridgeHandle, error)
	StageCreateWithOptions(key *Key, opts BridgeCreateOptions) (*BridgeHandle, error)

	// Get gets the BridgeHandle
	Get(key *Key) *BridgeHandle

	// Lists returns the lists of bridges in asterisk, optionally using the key for filtering.
	List(*Key) ([]*Key, error)

	// Data gets the bridge data
	Data(key *Key) (*BridgeData, error)

	// AddChannel adds a channel to the bridge
	AddChannel(key *Key, channelID string) error

	// AddChannelWithOptions adds a channel to a bridge, specifying additional options to be applied to that channel
	AddChannelWithOptions(key *Key, channelID string, options *BridgeAddChannelOptions) error

	// RemoveChannel removes a channel from the bridge
	RemoveChannel(key *Key, channelID string) error

	// Delete deletes the bridge
	Delete(key *Key) error

	// MOH plays music on hold
	MOH(key *Key, moh string) error

	// StopMOH stops music on hold
	StopMOH(key *Key) error

	// Play plays the media URI to the bridge
	Play(key *Key, playbackID string, mediaURI ...string) (*PlaybackHandle, error)
	PlayWithOptions(key *Key, playbackID string, opts BridgePlayOptions) (*PlaybackHandle, error)
	// PlayWithoutID uses POST /bridges/{bridgeId}/play and returns Asterisk's assigned ID.
	PlayWithoutID(key *Key, opts BridgePlayOptions) (*PlaybackHandle, error)
	// PlayOnCollection uses POST /bridges/{bridgeId}/play with an optional playbackId.
	PlayOnCollection(key *Key, playbackID string, opts BridgePlayOptions) (*PlaybackHandle, error)

	// StagePlay stages a `Play` operation and returns the `PlaybackHandle`
	// for invoking it.
	StagePlay(key *Key, playbackID string, mediaURI ...string) (*PlaybackHandle, error)
	StagePlayWithOptions(key *Key, playbackID string, opts BridgePlayOptions) (*PlaybackHandle, error)

	// Record records the bridge
	Record(key *Key, name string, opts *RecordingOptions) (*LiveRecordingHandle, error)

	// StageRecord stages a `Record` operation and returns the `PlaybackHandle`
	// for invoking it.
	StageRecord(key *Key, name string, opts *RecordingOptions) (*LiveRecordingHandle, error)

	// Subscribe subscribes the given bridge events events
	Subscribe(key *Key, n ...string) Subscription

	// VideoSource add Channel as Video-Source-ID at bridge
	VideoSource(key *Key, channelID string) error

	// VideoSourceDelete delete Video-Source-ID from bridge
	VideoSourceDelete(key *Key) error

	// GetVariable reads a bridge variable or function.
	GetVariable(key *Key, name string) (string, error)

	// SetVariable writes a bridge variable; reportEvents controls inclusion in bridge events.
	SetVariable(key *Key, name, value string, reportEvents *bool) error

	// GetVariables reads multiple bridge variables or functions.
	GetVariables(key *Key, names ...string) (map[string]any, error)

	// SetVariables writes multiple bridge variables or functions.
	SetVariables(key *Key, values map[string]BridgeVariableAssignment) error
}

// BridgeVariableAssignment is a value in the bulk bridge variable request.
// Without ReportEvents, Asterisk accepts the short string form. With it, the
// request uses an object so explicit false and true are both preserved.
type BridgeVariableAssignment struct {
	Value        string
	ReportEvents *bool
}

// BridgeCreateVariable is an object-valued variable on bridge creation.
type BridgeCreateVariable struct {
	Value        string `json:"value"`
	ReportEvents *bool  `json:"report_events,omitempty"`
}

// BridgeCreateOptions controls bridge creation and its initial variables.
type BridgeCreateOptions struct {
	Type      string
	Name      string
	Variables map[string]BridgeCreateVariable
}

// BridgePlayOptions controls playback on a bridge. A nil offset or skip value
// leaves Asterisk's default in effect; a pointer preserves an explicit zero.
type BridgePlayOptions struct {
	Media           []string
	AnnouncerFormat string
	Lang            string
	OffsetMS        *int
	SkipMS          *int
}

// VariableAssignment is the common single value format for channel and bridge
// bulk variable writes.
type VariableAssignment = BridgeVariableAssignment

func (value BridgeVariableAssignment) MarshalJSON() ([]byte, error) {
	if value.ReportEvents == nil {
		return json.Marshal(value.Value)
	}
	return json.Marshal(struct {
		Value        string `json:"value"`
		ReportEvents bool   `json:"report_events"`
	}{Value: value.Value, ReportEvents: *value.ReportEvents})
}

func (value *BridgeVariableAssignment) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		value.ReportEvents = nil
		return json.Unmarshal(data, &value.Value)
	}
	var object struct {
		Value        string `json:"value"`
		ReportEvents *bool  `json:"report_events"`
	}
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	value.Value, value.ReportEvents = object.Value, object.ReportEvents
	return nil
}

// BridgeData describes an Asterisk Bridge, the entity which merges media from
// one or more channels into a common audio output
type BridgeData struct {
	// Key is the cluster-unique identifier for this bridge
	Key *Key `json:"key"`

	ID         string   `json:"id"`           // Unique Id for this bridge
	Class      string   `json:"bridge_class"` // Class of the bridge
	Type       string   `json:"bridge_type"`  // Type of bridge (mixing, holding, dtmf_events, proxy_media)
	ChannelIDs []string `json:"channels"`     // List of pariticipating channel ids
	Creator    string   `json:"creator"`      // Creating entity of the bridge
	Name       string   `json:"name"`         // The name of the bridge
	Technology string   `json:"technology"`   // Name of the bridging technology
}

// BridgeAddChannelOptions describes additional options to be applied to a channel when it is joined to a bridge
type BridgeAddChannelOptions struct {
	// AbsorbDTMF indicates that DTMF coming from this channel will not be passed through to the bridge
	AbsorbDTMF bool

	// Mute indicates that the channel should be muted, preventing audio from it passing through to the bridge
	Mute bool

	// Role indicates the channel's role in the bridge
	Role string

	// InhibitConnectedLineUpdates suppresses presenting the new channel's
	// identity to existing bridge members. Nil leaves the option omitted.
	InhibitConnectedLineUpdates *bool
}

// Channels returns the list of channels found in the bridge
func (b *BridgeData) Channels() (list []*Key) {
	for _, id := range b.ChannelIDs {
		list = append(list, b.Key.New(ChannelKey, id))
	}

	return
}

// NewBridgeHandle creates a new bridge handle
func NewBridgeHandle(key *Key, b Bridge, exec func(bh *BridgeHandle) error) *BridgeHandle {
	return &BridgeHandle{
		key:  key,
		b:    b,
		exec: exec,
	}
}

// BridgeHandle is the handle to a bridge for performing operations
type BridgeHandle struct {
	key      *Key
	b        Bridge
	exec     func(bh *BridgeHandle) error
	executed bool
}

// ID returns the identifier for the bridge
func (bh *BridgeHandle) ID() string {
	return bh.key.ID
}

// Key returns the Key of the bridge
func (bh *BridgeHandle) Key() *Key {
	return bh.key
}

// GetVariable reads a variable from this bridge.
func (bh *BridgeHandle) GetVariable(name string) (string, error) {
	return bh.b.GetVariable(bh.key, name)
}

// SetVariable writes a variable on this bridge.
func (bh *BridgeHandle) SetVariable(name, value string, reportEvents *bool) error {
	return bh.b.SetVariable(bh.key, name, value, reportEvents)
}

// GetVariables reads multiple variables from this bridge.
func (bh *BridgeHandle) GetVariables(names ...string) (map[string]any, error) {
	return bh.b.GetVariables(bh.key, names...)
}

// SetVariables writes multiple variables on this bridge.
func (bh *BridgeHandle) SetVariables(values map[string]BridgeVariableAssignment) error {
	return bh.b.SetVariables(bh.key, values)
}

// Exec executes any staged operations attached on the bridge handle
func (bh *BridgeHandle) Exec() error {
	if !bh.executed {
		bh.executed = true
		if bh.exec != nil {
			err := bh.exec(bh)
			bh.exec = nil

			return err
		}
	}

	return nil
}

// AddChannel adds a channel to the bridge
func (bh *BridgeHandle) AddChannel(channelID string) error {
	return bh.b.AddChannel(bh.key, channelID)
}

// AddChannelWithOptions adds a channel to the bridge, specifying additional options
func (bh *BridgeHandle) AddChannelWithOptions(channelID string, options *BridgeAddChannelOptions) error {
	return bh.b.AddChannelWithOptions(bh.key, channelID, options)
}

// RemoveChannel removes a channel from the bridge
func (bh *BridgeHandle) RemoveChannel(channelID string) error {
	return bh.b.RemoveChannel(bh.key, channelID)
}

// Delete deletes the bridge
func (bh *BridgeHandle) Delete() (err error) {
	err = bh.b.Delete(bh.key)
	return
}

// Data gets the bridge data
func (bh *BridgeHandle) Data() (*BridgeData, error) {
	return bh.b.Data(bh.key)
}

// MOH requests that the given MusicOnHold class being played to the bridge
func (bh *BridgeHandle) MOH(class string) error {
	return bh.b.MOH(bh.key, class)
}

// StopMOH requests that any MusicOnHold which is being played to the bridge is stopped.
func (bh *BridgeHandle) StopMOH() error {
	return bh.b.StopMOH(bh.key)
}

// Play initiates playback of the specified media uri
// to the bridge, returning the Playback handle
func (bh *BridgeHandle) Play(id string, mediaURI ...string) (*PlaybackHandle, error) {
	return bh.b.Play(bh.key, id, mediaURI...)
}

func (bh *BridgeHandle) PlayWithOptions(id string, opts BridgePlayOptions) (*PlaybackHandle, error) {
	return bh.b.PlayWithOptions(bh.key, id, opts)
}

func (bh *BridgeHandle) PlayWithoutID(opts BridgePlayOptions) (*PlaybackHandle, error) {
	return bh.b.PlayWithoutID(bh.key, opts)
}

func (bh *BridgeHandle) PlayOnCollection(id string, opts BridgePlayOptions) (*PlaybackHandle, error) {
	return bh.b.PlayOnCollection(bh.key, id, opts)
}

// StagePlay stages a `Play` operation.
func (bh *BridgeHandle) StagePlay(id string, mediaURI ...string) (*PlaybackHandle, error) {
	return bh.b.StagePlay(bh.key, id, mediaURI...)
}

func (bh *BridgeHandle) StagePlayWithOptions(id string, opts BridgePlayOptions) (*PlaybackHandle, error) {
	return bh.b.StagePlayWithOptions(bh.key, id, opts)
}

// Record records the bridge to the given filename
func (bh *BridgeHandle) Record(name string, opts *RecordingOptions) (*LiveRecordingHandle, error) {
	return bh.b.Record(bh.key, name, opts)
}

// StageRecord stages a `Record` operation
func (bh *BridgeHandle) StageRecord(name string, opts *RecordingOptions) (*LiveRecordingHandle, error) {
	return bh.b.StageRecord(bh.key, name, opts)
}

// Subscribe creates a subscription to the list of events
func (bh *BridgeHandle) Subscribe(n ...string) Subscription {
	if bh == nil {
		return nil
	}

	return bh.b.Subscribe(bh.key, n...)
}

// VideoSource sets channel as Video-Source-Id in a multi-party mixing bridge
func (bh *BridgeHandle) VideoSource(channelID string) error {
	return bh.b.VideoSource(bh.key, channelID)
}

// VideoSourceDelete deletes Video-Source-Id in a multi-party mixing bridge
func (bh *BridgeHandle) VideoSourceDelete() error {
	return bh.b.VideoSourceDelete(bh.key)
}
