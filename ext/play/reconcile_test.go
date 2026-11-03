// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package play

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/stdbus"
)

type reconcilePlayback struct {
	ari.Playback
	bus   ari.Bus
	read  func() (*ari.PlaybackData, error)
	reads atomic.Int32
	stops atomic.Int32
}

func (p *reconcilePlayback) Subscribe(k *ari.Key, events ...string) ari.Subscription {
	return p.bus.Subscribe(k, events...)
}
func (p *reconcilePlayback) Data(*ari.Key) (*ari.PlaybackData, error) {
	p.reads.Add(1)
	return p.read()
}
func (p *reconcilePlayback) Stop(*ari.Key) error { p.stops.Add(1); return nil }
func (p *reconcilePlayback) finish(state string) {
	p.bus.Send(&ari.PlaybackFinished{EventData: ari.EventData{Type: ari.Events.PlaybackFinished}, Playback: ari.PlaybackData{ID: "test", State: state}})
}
func (p *reconcilePlayback) start() {
	p.bus.Send(&ari.PlaybackStarted{EventData: ari.EventData{Type: ari.Events.PlaybackStarted}, Playback: ari.PlaybackData{ID: "test", State: "playing"}})
}
func newReconcilePlayback(t *testing.T) (*reconcilePlayback, *ari.PlaybackHandle) {
	p := &reconcilePlayback{bus: stdbus.New()}
	t.Cleanup(p.bus.Close)
	h := ari.NewPlaybackHandle(ari.NewKey(ari.PlaybackKey, "test"), p, nil)
	return p, h
}
func checkPlaybackResult(t *testing.T, p *reconcilePlayback, got, want Status, err error, wantErr bool) {
	t.Helper()
	if got != want || (err != nil) != wantErr {
		t.Fatalf("status=%v err=%v; want %v, error=%v", got, err, want, wantErr)
	}
	if p.stops.Load() != 1 {
		t.Fatalf("Stop calls=%d", p.stops.Load())
	}
}

func TestReconcileMissingStart(t *testing.T) {
	for _, state := range []string{"playing", "continuing"} {
		t.Run(state, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p, h := newReconcilePlayback(t)
				p.read = func() (*ari.PlaybackData, error) { return &ari.PlaybackData{State: state}, nil }
				finished := make(chan struct{})
				go func() {
					defer close(finished)
					time.Sleep(2 * time.Second)
					if p.stops.Load() != 0 {
						t.Error("active audio was stopped")
					}
					p.finish("done")
				}()
				start := time.Now()
				status, err := playStaged(context.Background(), h, 10*time.Millisecond)
				<-finished
				checkPlaybackResult(t, p, status, Finished, err, false)
				if time.Since(start) != 2*time.Second || p.reads.Load() == 0 {
					t.Fatal("did not preserve active playback through missing start")
				}
			})
		})
	}
}

type playbackGoneError struct{}

func (playbackGoneError) Error() string { return "not found" }
func (playbackGoneError) Code() int     { return 404 }

func TestReconcileTerminalStates(t *testing.T) {
	for _, tc := range []struct {
		name, state string
		readErr     error
		want        Status
		wantErr     bool
	}{
		{"done", "done", nil, Finished, false},
		{"failed", "failed", nil, Failed, true},
		{"queued", "queued", nil, Timeout, true},
		{"paused", "paused", nil, Failed, true},
		{"unknown", "", nil, Failed, true},
		{"gone", "", playbackGoneError{}, Failed, true},
		{"readFailure", "", errors.New("unavailable"), Failed, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p, h := newReconcilePlayback(t)
				p.read = func() (*ari.PlaybackData, error) { return &ari.PlaybackData{State: tc.state}, tc.readErr }
				status, err := playStaged(context.Background(), h, time.Millisecond)
				checkPlaybackResult(t, p, status, tc.want, err, tc.wantErr)
			})
		})
	}
}

func TestReconcileBothEventsMissing(t *testing.T) {
	for _, gone := range []bool{false, true} {
		t.Run(map[bool]string{false: "doneState", true: "goneIsNotSuccess"}[gone], func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p, h := newReconcilePlayback(t)
				p.read = func() (*ari.PlaybackData, error) {
					switch p.reads.Load() {
					case 1:
						return &ari.PlaybackData{State: "playing"}, nil
					case 2:
						return nil, errors.New("transient read failure")
					default:
						if gone {
							return nil, playbackGoneError{}
						}
						return &ari.PlaybackData{State: "done"}, nil
					}
				}
				status, err := playStaged(context.Background(), h, time.Millisecond)
				want := Finished
				if gone {
					want = Failed
				}
				checkPlaybackResult(t, p, status, want, err, gone)
				if p.reads.Load() != 3 {
					t.Fatalf("reads=%d", p.reads.Load())
				}
			})
		})
	}
}

func TestReconcileEventPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		cancel, lateStart, failed bool
	}{
		{name: "finishBefore404"}, {name: "failedFinish", failed: true}, {name: "cancelBeforeFinish", cancel: true}, {name: "lateStartBeforeQueued", lateStart: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p, h := newReconcilePlayback(t)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				p.read = func() (*ari.PlaybackData, error) {
					if tc.cancel {
						cancel()
					}
					if tc.lateStart {
						p.start()
						return &ari.PlaybackData{State: "queued"}, nil
					}
					state := "done"
					if tc.failed {
						state = "failed"
					}
					p.finish(state)
					return nil, playbackGoneError{}
				}
				if tc.lateStart {
					go func() { time.Sleep(100 * time.Millisecond); p.finish("done") }()
				}
				status, err := playStaged(ctx, h, time.Millisecond)
				want := Finished
				if tc.cancel {
					want = Cancelled
				}
				if tc.failed {
					want = Failed
				}
				checkPlaybackResult(t, p, status, want, err, tc.failed)
			})
		})
	}
}

func TestReconcileSlowReadDoesNotBlockEventsOrCancellation(t *testing.T) {
	for _, cancelCall := range []bool{false, true} {
		t.Run(map[bool]string{false: "finish", true: "cancel"}[cancelCall], func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p, h := newReconcilePlayback(t)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				release := make(chan struct{})
				entered := make(chan struct{})
				p.read = func() (*ari.PlaybackData, error) {
					close(entered)
					<-release
					return &ari.PlaybackData{State: "playing"}, nil
				}
				go func() {
					<-entered
					if cancelCall {
						cancel()
					} else {
						p.finish("done")
					}
				}()
				status, err := playStaged(ctx, h, time.Millisecond)
				close(release)
				want := Finished
				if cancelCall {
					want = Cancelled
				}
				checkPlaybackResult(t, p, status, want, err, false)
				synctest.Wait()
			})
		})
	}
}

func TestReconcileClosedSubscription(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p, h := newReconcilePlayback(t)
		p.read = func() (*ari.PlaybackData, error) { p.bus.Close(); return &ari.PlaybackData{State: "playing"}, nil }
		status, err := playStaged(context.Background(), h, time.Millisecond)
		checkPlaybackResult(t, p, status, Failed, err, true)
	})
}

type reconcilePlayer struct {
	ari.Player
	playback *reconcilePlayback
	handle   *ari.PlaybackHandle
}

func (p *reconcilePlayer) StagePlay(string, ...string) (*ari.PlaybackHandle, error) {
	return p.handle, nil
}
func (p *reconcilePlayer) Subscribe(events ...string) ari.Subscription {
	return p.playback.bus.Subscribe(nil, events...)
}

func TestReconcileSessionInterruption(t *testing.T) {
	for _, mode := range []string{"dtmf", "sessionLimit"} {
		t.Run(mode, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p, h := newReconcilePlayback(t)
				p.read = func() (*ari.PlaybackData, error) { return &ari.PlaybackData{State: "playing"}, nil }
				player := &reconcilePlayer{playback: p, handle: h}
				opts := NewDefaultOptions()
				if err := opts.ApplyOptions(URI("sound:test"), PlaybackStartTimeout(time.Millisecond)); err != nil {
					t.Fatal(err)
				}
				if mode == "sessionLimit" {
					opts.maxPlaybackTime = 100 * time.Millisecond
				}
				start := time.Now()
				session := opts.Play(context.Background(), player)
				if mode == "dtmf" {
					go func() {
						time.Sleep(100 * time.Millisecond)
						p.bus.Send(&ari.ChannelDtmfReceived{EventData: ari.EventData{Type: ari.Events.ChannelDtmfReceived}, Channel: ari.ChannelData{ID: "customer"}, Digit: "5"})
					}()
				}
				<-session.Done()
				synctest.Wait()
				if time.Since(start) != 100*time.Millisecond || p.stops.Load() != 1 || p.reads.Load() != 1 {
					t.Fatalf("elapsed=%s stops=%d reads=%d", time.Since(start), p.stops.Load(), p.reads.Load())
				}
				if mode == "dtmf" {
					result, err := session.Result()
					if err != nil || result.DTMF != "5" {
						t.Fatalf("result=%+v err=%v", result, err)
					}
				}
			})
		})
	}
}
