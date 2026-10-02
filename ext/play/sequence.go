package play

import (
	"context"
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

	select {
	case <-ctx.Done():
		return Cancelled, nil
	case <-started.Events():
	case <-finished.Events():
		return Finished, nil
	case <-time.After(timeout):
		return Timeout, eris.New("timeout waiting for playback to start")
	}

	// Wait for playback to complete
	select {
	case <-ctx.Done():
		return Cancelled, nil
	case <-finished.Events():
		return Finished, nil
	}
}
