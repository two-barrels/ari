// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package stdbus

import (
	"sync"

	"github.com/two-barrels/ari/v6"
)

// subscriptionEventBufferSize defines the number of events that each
// subscription will queue before accepting more events.
var subscriptionEventBufferSize = 100

// bus is an event bus for ARI events.  It receives and
// redistributes events based on a subscription
// model.
type bus struct {
	subs []*subscription // The list of subscriptions

	rwMux sync.RWMutex

	closed    bool
	closeOnce sync.Once
}

// New creates and returns the event bus.
func New() ari.Bus {
	b := &bus{
		subs: []*subscription{},
	}

	return b
}

// Close closes out all subscriptions in the bus. Concurrent and repeated
// calls wait for the initial shutdown to finish.
func (b *bus) Close() {
	b.closeOnce.Do(func() {
		b.rwMux.Lock()
		b.closed = true
		subs := b.subs
		b.subs = nil
		b.rwMux.Unlock()

		// Cancel reacquires the bus lock through remove, so it must run
		// after the subscription list has been detached and unlocked.
		for _, s := range subs {
			s.Cancel()
		}
	})
}

// Send sends the message to the bus
func (b *bus) Send(e ari.Event) {
	var matched bool
	keys := e.Keys()
	if len(keys) == 0 {
		// Some ARI events describe the application itself and have no resource.
		// They still belong on application-wide or all-event subscriptions.
		keys = ari.Keys{e.Key("", "")}
	}

	b.rwMux.RLock()

	// Disseminate the message to the subscribers
	for _, s := range b.subs {
		matched = false
		for _, k := range keys {
			if k.Kind == "" && k.ID == "" && s.key != nil && (s.key.Kind != "" || s.key.ID != "") {
				continue
			}
			if matched {
				break
			}

			if s.key.Match(k) {
				matched = true

				for _, topic := range s.events {
					if topic == e.GetType() || topic == ari.Events.All {
						select {
						case s.C <- e:
						default: // never block
						}
					}
				}
			}
		}
	}

	b.rwMux.RUnlock()
}

// Subscribe returns a subscription to the given list
// of event types. After the bus is closed, it returns an already-cancelled
// subscription whose Events channel is closed.
func (b *bus) Subscribe(key *ari.Key, eTypes ...string) ari.Subscription {
	s := newSubscription(b, key, eTypes...)
	if !b.add(s) {
		s.Cancel()
	}

	return s
}

// add appends a new subscription to the bus
func (b *bus) add(s *subscription) bool {
	b.rwMux.Lock()
	defer b.rwMux.Unlock()
	if b.closed {
		return false
	}
	b.subs = append(b.subs, s)
	return true
}

// remove deletes the given subscription from the bus
func (b *bus) remove(s *subscription) {
	b.rwMux.Lock()

	for i, si := range b.subs {
		if s == si {
			// Subs are pointers, so we have to explicitly remove them
			// to prevent memory leaks
			b.subs[i] = b.subs[len(b.subs)-1] // replace the current with the end
			b.subs[len(b.subs)-1] = nil       // remove the end
			b.subs = b.subs[:len(b.subs)-1]   // lop off the end

			break
		}
	}

	b.rwMux.Unlock()
}

// A Subscription is a wrapped channel for receiving
// events from the ARI event bus.
type subscription struct {
	key    *ari.Key
	b      *bus     // reference to the event bus
	events []string // list of events to listen for

	mu     sync.Mutex
	closed bool           // channel closure protection flag
	C      chan ari.Event // channel for sending events to the subscriber
}

// newSubscription creates a new, unattached subscription
func newSubscription(b *bus, key *ari.Key, eTypes ...string) *subscription {
	return &subscription{
		key:    key,
		b:      b,
		events: eTypes,
		C:      make(chan ari.Event, subscriptionEventBufferSize),
	}
}

// Events returns the events channel
func (s *subscription) Events() <-chan ari.Event {
	return s.C
}

// Cancel cancels the subscription and removes it from
// the event bus.
func (s *subscription) Cancel() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true

	// Remove the subscription from the bus
	if s.b != nil {
		s.b.remove(s)
	}

	// Close the subscription's deliver channel
	if s.C != nil {
		close(s.C)
	}
}
