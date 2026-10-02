package stdbus

import (
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestApplicationWideSubscriptionReceivesKeylessEvent(t *testing.T) {
	b := New()
	all := b.Subscribe(nil, ari.Events.All)
	defer all.Cancel()
	channel := b.Subscribe(ari.NewKey(ari.ChannelKey, "channel-1"), ari.Events.All)
	defer channel.Cancel()

	b.Send(&ari.ApplicationReplaced{EventData: ari.EventData{Application: "demo", Type: ari.Events.ApplicationReplaced}})
	select {
	case event := <-all.Events():
		if event.GetType() != ari.Events.ApplicationReplaced {
			t.Fatalf("received %q, want ApplicationReplaced", event.GetType())
		}
	default:
		t.Fatal("application-wide subscription did not receive keyless event")
	}
	select {
	case <-channel.Events():
		t.Fatal("resource-specific subscription received keyless event")
	default:
	}
}
