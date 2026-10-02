// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package stdbus

import (
	"sync"
	"testing"
	"time"

	"github.com/two-barrels/ari/v6"
)

func assertSubscriptionClosed(t *testing.T, sub ari.Subscription) {
	t.Helper()
	// Buffered events may remain after cancellation, but the channel must
	// be closed once those events have been drained.
	for {
		select {
		case _, ok := <-sub.Events():
			if !ok {
				return
			}
		default:
			t.Fatal("subscription channel is still open")
		}
	}
}

func TestSubscribeAfterClose(t *testing.T) {
	b := New()
	b.Close()
	sub := b.Subscribe(nil, ari.Events.All)
	assertSubscriptionClosed(t, sub)
	sub.Cancel()
	b.Send(dtmfTestEvent)
	b.Close()
}

func TestConcurrentBusShutdown(t *testing.T) {
	b := New().(*bus)
	var retained []ari.Subscription
	for i := 0; i < 16; i++ {
		retained = append(retained, b.Subscribe(nil, ari.Events.All))
	}
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			sub := b.Subscribe(nil, ari.Events.All)
			<-start
			for j := 0; j < 100; j++ {
				b.Send(dtmfTestEvent)
				sub.Cancel()
				sub = b.Subscribe(nil, ari.Events.All)
			}
			b.Close()
			assertSubscriptionClosed(t, sub)
		}()
	}
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			b.Close()
			for _, sub := range retained {
				assertSubscriptionClosed(t, sub)
			}
			assertSubscriptionClosed(t, b.Subscribe(nil, ari.Events.All))
		}()
	}
	close(start)
	done := make(chan struct{})
	go func() {
		workers.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent shutdown deadlocked")
	}
	b.rwMux.RLock()
	defer b.rwMux.RUnlock()
	if !b.closed || len(b.subs) != 0 {
		t.Fatalf("shutdown left closed=%v, subscriptions=%d", b.closed, len(b.subs))
	}
}

func TestConcurrentSubscriptionCancel(t *testing.T) {
	b := New()
	defer b.Close()
	sub := b.Subscribe(nil, ari.Events.All)
	var workers sync.WaitGroup
	for i := 0; i < 16; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			sub.Cancel()
			assertSubscriptionClosed(t, sub)
		}()
	}
	workers.Wait()
}
