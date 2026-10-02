package native

import (
	"net/url"
	"reflect"
	"testing"
)

func TestEventWebsocketQuery(t *testing.T) {
	tests := []struct {
		name         string
		subscribeAll bool
		want         url.Values
	}{
		{"application only", false, url.Values{"app": {"demo+app"}}},
		{"subscribe all", true, url.Values{"app": {"demo+app"}, "subscribeAll": {"true"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := New(&Options{Application: "demo+app", WebsocketURL: "ws://asterisk.test/ari/events", SubscribeAll: test.subscribeAll})
			if err := c.createWSConfig(); err != nil {
				t.Fatal(err)
			}
			if c.WSConfig.Location.Scheme != "ws" || c.WSConfig.Location.EscapedPath() != "/ari/events" || !reflect.DeepEqual(c.WSConfig.Location.Query(), test.want) {
				t.Errorf("websocket URL=%s, want query=%s", c.WSConfig.Location, test.want.Encode())
			}
		})
	}
}
