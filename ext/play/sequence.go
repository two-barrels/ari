// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package play

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rotisserie/eris"

	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/rid"
)

// sequence represents an audio sequence playback session
type sequence struct {
	ctx    context.Context
	cancel context.CancelFunc
	s      *playSession

	done chan struct{}
}

func (s *sequence) Done() <-chan struct{} {
	return s.done
}

func (s *sequence) Stop() {
	s.cancel()
}

func newSequence(ctx context.Context, s *playSession) *sequence {
	ctx, cancel := context.WithCancel(ctx)
	return &sequence{
		ctx:    ctx,
		cancel: cancel,
		s:      s,
		done:   make(chan struct{}),
	}
}

func (s *sequence) Play(p ari.Player, playbackCounter int) {
	defer s.cancel()
	defer close(s.done)
	if s.ctx.Err() != nil {
		s.s.result.Status = Cancelled
		return
	}

	if playbackCounter > 0 && !s.s.o.invalidPrependUriList.Empty() {
		for u := s.s.o.invalidPrependUriList.First(); u != ""; u = s.s.o.invalidPrependUriList.Next() {
			if s.ctx.Err() != nil {
				s.s.result.Status = Cancelled
				return
			}
			pb, err := p.StagePlay(rid.New(rid.Playback), u)
			if err != nil {
				s.s.result.Status = Failed
				s.s.result.Error = eris.Wrap(err, "failed to stage playback")

				return
			}

			s.s.result.Status, err = playStaged(s.ctx, pb, s.s.o.playbackStartTimeout)
			if err != nil {
				s.s.result.Error = eris.Wrap(err, "failure in playback")

				return
			}
		}
	}

	for u := s.s.o.uriList.First(); u != ""; u = s.s.o.uriList.Next() {
		if s.ctx.Err() != nil {
			s.s.result.Status = Cancelled
			return
		}
		pb, err := p.StagePlay(rid.New(rid.Playback), u)
		if err != nil {
			s.s.result.Status = Failed
			s.s.result.Error = eris.Wrap(err, "failed to stage playback")

			return
		}

		s.s.result.Status, err = playStaged(s.ctx, pb, s.s.o.playbackStartTimeout)
		if err != nil {
			s.s.result.Error = eris.Wrap(err, "failure in playback")

			return
		}
	}
}

// playStaged executes a staged playback, waiting for its completion
func playStaged(ctx context.Context, h *ari.PlaybackHandle, timeout time.Duration) (Status, error) {
	if ctx.Err() != nil {
		return Cancelled, nil
	}
	started := h.Subscribe(ari.Events.PlaybackStarted)
	defer started.Cancel()

	finished := h.Subscribe(ari.Events.PlaybackFinished)
	defer finished.Cancel()

	if timeout == 0 {
		timeout = DefaultPlaybackStartTimeout
	}
	if ctx.Err() != nil {
		return Cancelled, nil
	}

	if err := h.Exec(); err != nil {
		return Failed, eris.Wrap(err, "failed to start playback")
	}

	defer h.Stop() // nolint: errcheck

	// A missing event is not evidence that Asterisk failed to start audio.
	// Reconcile only this exceptional path; normal playback remains event driven.
	startTimer := time.NewTimer(timeout)
	defer startTimer.Stop()
	var retry *time.Timer
	defer func() {
		if retry != nil {
			retry.Stop()
		}
	}()
	probeTimer := startTimer.C
	startedEvents, finishedEvents := started.Events(), finished.Events()
	confirmed := false
	type probeResult struct {
		data *ari.PlaybackData
		err  error
	}
	var probe <-chan probeResult

	// Queued events take precedence over a potentially stale REST snapshot.
	// The channel is already scoped to the playback by Subscribe.
	drain := func() (Status, error, bool) {
		if ctx.Err() != nil {
			return Cancelled, nil, true
		}
		select {
		case event, ok := <-finishedEvents:
			status, err := playbackFinished(event, ok)
			return status, err, true
		default:
		}
		select {
		case _, ok := <-startedEvents:
			if !ok {
				return Failed, eris.New("playback event subscription closed"), true
			}
			confirmed = true
		default:
		}
		return InProgress, nil, false
	}
	for {
		select {
		case <-ctx.Done():
			return Cancelled, nil
		case event, ok := <-finishedEvents:
			if ctx.Err() != nil {
				return Cancelled, nil
			}
			return playbackFinished(event, ok)
		case _, ok := <-startedEvents:
			if !ok {
				return Failed, eris.New("playback event subscription closed")
			}
			confirmed = true
			// If no probe has started, retain the ordinary event-only path.
			if retry == nil && probe == nil {
				startTimer.Stop()
				probeTimer = nil
			}
		case <-probeTimer:
			probeTimer = nil
			if status, err, done := drain(); done {
				return status, err
			}
			if confirmed && retry == nil {
				continue
			}
			// Data has no context argument. Keep cancellation and playback
			// events responsive during the read, with at most one read in flight.
			// The client's transport timeout bounds the outstanding request.
			result := make(chan probeResult, 1)
			probe = result
			go func() {
				data, err := h.Data()
				result <- probeResult{data: data, err: err}
			}()
		case result := <-probe:
			probe = nil
			if status, err, done := drain(); done {
				return status, err
			}
			if result.err != nil {
				var coded interface{ Code() int }
				if errors.As(result.err, &coded) && coded.Code() == http.StatusNotFound {
					return Failed, eris.Wrap(result.err, "playback disappeared without a completion event")
				}
				if !confirmed {
					return Failed, eris.Wrap(result.err, "unable to read playback state")
				}
			} else {
				state := ""
				if result.data != nil {
					state = result.data.State
				}
				switch state {
				case "playing", "continuing":
					confirmed = true
				case "done":
					return Finished, nil
				case "failed":
					return Failed, eris.New("playback failed")
				case "queued":
					if !confirmed {
						return Timeout, eris.New("timeout waiting for playback to start")
					}
				default:
					if !confirmed {
						return Failed, eris.Errorf("unexpected playback state %q", state)
					}
				}
			}
			// Continue reconciling if completion events are also lost. The
			// session's time limit and caller context still bound the wait.
			retry = time.NewTimer(500 * time.Millisecond)
			probeTimer = retry.C
		}
	}
}

func playbackFinished(event ari.Event, open bool) (Status, error) {
	if !open {
		return Failed, eris.New("playback event subscription closed")
	}
	if finished, ok := event.(*ari.PlaybackFinished); ok && finished.Playback.State == "failed" {
		return Failed, eris.New("playback failed")
	}
	return Finished, nil
}
