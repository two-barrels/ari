// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

// Compatibility methods for checked-in mocks. Remove this file when mockery
// regenerates Application.go and Asterisk.go from the updated interfaces.
package arimocks

import (
	"context"
	"github.com/two-barrels/ari/v6"
	"github.com/stretchr/testify/mock"
)

var (
	_ ari.Application     = (*Application)(nil)
	_ ari.Asterisk        = (*Asterisk)(nil)
	_ ari.Bridge          = (*Bridge)(nil)
	_ ari.Channel         = (*Channel)(nil)
	_ ari.Endpoint        = (*Endpoint)(nil)
	_ ari.StoredRecording = (*StoredRecording)(nil)
	_ ari.TextMessage     = (*TextMessage)(nil)
)

// FilterEvents supports the newer application filter method on the checked-in mock.
func (m *Application) FilterEvents(key *ari.Key, filter *ari.ApplicationEventFilter) (*ari.ApplicationData, error) {
	ret := m.Called(key, filter)
	if fn, ok := ret.Get(0).(func(*ari.Key, *ari.ApplicationEventFilter) (*ari.ApplicationData, error)); ok {
		return fn(key, filter)
	}
	var data *ari.ApplicationData
	if ret.Get(0) != nil {
		data = ret.Get(0).(*ari.ApplicationData)
	}
	return data, ret.Error(1)
}

func (e *Application_Expecter) FilterEvents(key, filter interface{}) *mock.Call {
	return e.mock.On("FilterEvents", key, filter)
}

func (m *Application) ClaimChannel(key *ari.Key, channelID string) error {
	return m.Called(key, channelID).Error(0)
}

// Ping supports the newer Asterisk ping method on the checked-in mock.
func (m *Asterisk) Ping(key *ari.Key) (*ari.AsteriskPing, error) {
	ret := m.Called(key)
	if fn, ok := ret.Get(0).(func(*ari.Key) (*ari.AsteriskPing, error)); ok {
		return fn(key)
	}
	var data *ari.AsteriskPing
	if ret.Get(0) != nil {
		data = ret.Get(0).(*ari.AsteriskPing)
	}
	return data, ret.Error(1)
}

func (m *Asterisk) InfoWithOptions(key *ari.Key, opts ari.AsteriskInfoOptions) (*ari.AsteriskInfo, error) {
	ret := m.Called(key, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.AsteriskInfo), ret.Error(1)
}

func (e *Asterisk_Expecter) Ping(key interface{}) *mock.Call {
	return e.mock.On("Ping", key)
}

func (m *Bridge) GetVariable(key *ari.Key, name string) (string, error) {
	ret := m.Called(key, name)
	return ret.String(0), ret.Error(1)
}

func (m *Bridge) SetVariable(key *ari.Key, name, value string, reportEvents *bool) error {
	return m.Called(key, name, value, reportEvents).Error(0)
}

func (m *Bridge) GetVariables(key *ari.Key, names ...string) (map[string]any, error) {
	ret := m.Called(key, names)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(map[string]any), ret.Error(1)
}

func (m *Bridge) SetVariables(key *ari.Key, values map[string]ari.BridgeVariableAssignment) error {
	return m.Called(key, values).Error(0)
}

func (m *Bridge) CreateWithOptions(key *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	ret := m.Called(key, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.BridgeHandle), ret.Error(1)
}

func (m *Bridge) CreateWithoutID(key *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	ret := m.Called(key, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.BridgeHandle), ret.Error(1)
}

func (m *Bridge) CreateOnCollection(key *ari.Key, id string, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	ret := m.Called(key, id, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.BridgeHandle), ret.Error(1)
}

func (m *Bridge) PlayOnCollection(key *ari.Key, id string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	ret := m.Called(key, id, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.PlaybackHandle), ret.Error(1)
}

func (m *Bridge) PlayWithoutID(key *ari.Key, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	ret := m.Called(key, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.PlaybackHandle), ret.Error(1)
}

func (m *Bridge) StageCreateWithOptions(key *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	ret := m.Called(key, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.BridgeHandle), ret.Error(1)
}

func (m *Bridge) PlayWithOptions(key *ari.Key, playbackID string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	ret := m.Called(key, playbackID, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.PlaybackHandle), ret.Error(1)
}

func (m *Bridge) StagePlayWithOptions(key *ari.Key, playbackID string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	ret := m.Called(key, playbackID, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.PlaybackHandle), ret.Error(1)
}

func (e *Bridge_Expecter) GetVariable(key, name interface{}) *mock.Call {
	return e.mock.On("GetVariable", key, name)
}
func (e *Bridge_Expecter) SetVariable(key, name, value, reportEvents interface{}) *mock.Call {
	return e.mock.On("SetVariable", key, name, value, reportEvents)
}
func (e *Bridge_Expecter) GetVariables(key, names interface{}) *mock.Call {
	return e.mock.On("GetVariables", key, names)
}
func (e *Bridge_Expecter) SetVariables(key, values interface{}) *mock.Call {
	return e.mock.On("SetVariables", key, values)
}

func (m *Channel) GetVariables(key *ari.Key, names ...string) (map[string]any, error) {
	ret := m.Called(key, names)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(map[string]any), ret.Error(1)
}

func (m *Channel) SetVariables(key *ari.Key, values map[string]ari.VariableAssignment) error {
	return m.Called(key, values).Error(0)
}

func (m *Channel) ContinueWithOptions(key *ari.Key, opts ari.ChannelContinueOptions) error {
	return m.Called(key, opts).Error(0)
}

func (m *Channel) HangupWithOptions(key *ari.Key, opts ari.ChannelHangupOptions) error {
	return m.Called(key, opts).Error(0)
}

func (m *Channel) SetVariableWithOptions(key *ari.Key, name, value string, opts *ari.ChannelVariableSetOptions) error {
	return m.Called(key, name, value, opts).Error(0)
}

func (m *Channel) PlayWithOptions(key *ari.Key, playbackID string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	ret := m.Called(key, playbackID, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.PlaybackHandle), ret.Error(1)
}

func (m *Channel) OriginateWithID(key *ari.Key, req ari.OriginateRequest) (*ari.ChannelHandle, error) {
	ret := m.Called(key, req)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.ChannelHandle), ret.Error(1)
}

func (m *Channel) PlayWithoutID(key *ari.Key, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	ret := m.Called(key, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.PlaybackHandle), ret.Error(1)
}

func (m *Channel) SnoopWithoutID(key *ari.Key, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	ret := m.Called(key, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.ChannelHandle), ret.Error(1)
}

func (m *Channel) PlayOnCollection(key *ari.Key, id string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	ret := m.Called(key, id, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.PlaybackHandle), ret.Error(1)
}

func (m *Channel) SnoopOnCollection(key *ari.Key, id string, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	ret := m.Called(key, id, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.ChannelHandle), ret.Error(1)
}

func (m *Channel) StagePlayWithOptions(key *ari.Key, playbackID string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	ret := m.Called(key, playbackID, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.PlaybackHandle), ret.Error(1)
}

func (m *Channel) Redirect(key *ari.Key, endpoint string) error {
	return m.Called(key, endpoint).Error(0)
}

func (m *Channel) Progress(key *ari.Key) error {
	return m.Called(key).Error(0)
}

func (m *Channel) TransferProgress(key *ari.Key, state string) error {
	return m.Called(key, state).Error(0)
}

func (m *Channel) RTPStatistics(key *ari.Key) (*ari.RTPStats, error) {
	ret := m.Called(key)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.RTPStats), ret.Error(1)
}

func (m *Endpoint) Refer(key *ari.Key, opts ari.EndpointReferOptions) error {
	return m.Called(key, opts).Error(0)
}

func (m *Endpoint) ReferToEndpoint(key *ari.Key, opts ari.EndpointReferOptions) error {
	return m.Called(key, opts).Error(0)
}

func (m *StoredRecording) File(ctx context.Context, key *ari.Key) (*ari.RecordingFile, error) {
	ret := m.Called(ctx, key)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*ari.RecordingFile), ret.Error(1)
}

func (m *TextMessage) SendWithKey(key *ari.Key, from, body string, vars map[string]string) error {
	return m.Called(key, from, body, vars).Error(0)
}

func (m *TextMessage) SendByURIWithKey(key *ari.Key, from, to, body string, vars map[string]string) error {
	return m.Called(key, from, to, body, vars).Error(0)
}
