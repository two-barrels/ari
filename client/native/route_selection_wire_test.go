package native

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestExplicitAlternateRoutesWire(t *testing.T) {
	zero := 0
	tests := []struct {
		name, path, responseID, kind string
		query                        url.Values
		body                         map[string]any
		absent                       string
		call                         func(*Client) (string, error)
	}{
		{
			name: "bridge create without ID", path: "/ari/bridges", responseID: "server-bridge", kind: "bridge",
			query: url.Values{"type": {"mixing"}, "name": {"support+one"}}, absent: "bridgeId",
			call: func(c *Client) (string, error) {
				h, err := c.Bridge().CreateWithoutID(nil, ari.BridgeCreateOptions{Type: "mixing", Name: "support+one"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "bridge create collection with ID", path: "/ari/bridges", responseID: "bridge+9", kind: "bridge",
			query: url.Values{"bridgeId": {"bridge+9"}, "type": {"mixing"}, "name": {"support+two"}}, body: map[string]any{},
			call: func(c *Client) (string, error) {
				h, err := c.Bridge().CreateOnCollection(nil, "bridge+9", ari.BridgeCreateOptions{Type: "mixing", Name: "support+two"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "bridge create with path ID and omitted options", path: "/ari/bridges/bridge-empty", responseID: "bridge-empty", kind: "bridge",
			query: url.Values{}, body: map[string]any{}, absent: "variables",
			call: func(c *Client) (string, error) {
				h, err := c.Bridge().CreateWithOptions(ari.NewKey(ari.BridgeKey, "bridge-empty"), ari.BridgeCreateOptions{})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "bridge play without ID", path: "/ari/bridges/bridge-1/play", responseID: "server-playback", kind: "playback",
			query: url.Values{"media": {"sound:one"}, "announcer_format": {"slin16"}, "lang": {"en"}, "offsetms": {"0"}, "skipms": {"4500"}}, absent: "playbackId",
			call: func(c *Client) (string, error) {
				skip := 4500
				h, err := c.Bridge().PlayWithoutID(ari.NewKey(ari.BridgeKey, "bridge-1"), ari.BridgePlayOptions{Media: []string{"sound:one"}, AnnouncerFormat: "slin16", Lang: "en", OffsetMS: &zero, SkipMS: &skip})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "bridge play collection with ID", path: "/ari/bridges/bridge-1/play", responseID: "play+9", kind: "playback",
			query: url.Values{"media": {"sound:one"}, "playbackId": {"play+9"}},
			call: func(c *Client) (string, error) {
				h, err := c.Bridge().PlayOnCollection(ari.NewKey(ari.BridgeKey, "bridge-1"), "play+9", ari.BridgePlayOptions{Media: []string{"sound:one"}})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "channel originate with path ID", path: "/ari/channels/channel-9", responseID: "channel-9", kind: "channel",
			query: url.Values{}, body: map[string]any{"channelId": "channel-9", "endpoint": "PJSIP/alice", "app": "demo"},
			call: func(c *Client) (string, error) {
				h, err := c.Channel().OriginateWithID(nil, ari.OriginateRequest{ChannelID: "channel-9", Endpoint: "PJSIP/alice", App: "demo"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "channel play without ID", path: "/ari/channels/channel-1/play", responseID: "server-playback", kind: "playback",
			query: url.Values{"media": {"sound:two"}, "lang": {"en+US"}, "offsetms": {"300"}, "skipms": {"0"}}, absent: "playbackId",
			call: func(c *Client) (string, error) {
				offset := 300
				h, err := c.Channel().PlayWithoutID(ari.NewKey(ari.ChannelKey, "channel-1"), ari.ChannelPlayOptions{Media: []string{"sound:two"}, Lang: "en+US", OffsetMS: &offset, SkipMS: &zero})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "channel play collection with ID", path: "/ari/channels/channel-1/play", responseID: "play+9", kind: "playback",
			query: url.Values{"media": {"sound:two"}, "playbackId": {"play+9"}},
			call: func(c *Client) (string, error) {
				h, err := c.Channel().PlayOnCollection(ari.NewKey(ari.ChannelKey, "channel-1"), "play+9", ari.ChannelPlayOptions{Media: []string{"sound:two"}})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "channel snoop without ID", path: "/ari/channels/channel-1/snoop", responseID: "server-snoop", kind: "channel",
			query: url.Values{"app": {"demo"}, "appArgs": {"one,two"}, "spy": {"in"}, "whisper": {"out"}}, absent: "snoopId",
			call: func(c *Client) (string, error) {
				h, err := c.Channel().SnoopWithoutID(ari.NewKey(ari.ChannelKey, "channel-1"), &ari.SnoopOptions{App: "demo", AppArgs: "one,two", Spy: ari.DirectionIn, Whisper: ari.DirectionOut})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "channel snoop collection with ID", path: "/ari/channels/channel-1/snoop", responseID: "snoop+9", kind: "channel",
			query: url.Values{"snoopId": {"snoop+9"}, "app": {"demo"}},
			call: func(c *Client) (string, error) {
				h, err := c.Channel().SnoopOnCollection(ari.NewKey(ari.ChannelKey, "channel-1"), "snoop+9", &ari.SnoopOptions{App: "demo"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
		{
			name: "channel snoop path ID", path: "/ari/channels/channel-1/snoop/snoop-9", responseID: "snoop-9", kind: "channel",
			query: url.Values{"app": {"demo"}, "appArgs": {"one,two"}, "spy": {"in"}, "whisper": {"out"}}, absent: "snoopId",
			call: func(c *Client) (string, error) {
				h, err := c.Channel().Snoop(ari.NewKey(ari.ChannelKey, "channel-1"), "snoop-9", &ari.SnoopOptions{App: "demo", AppArgs: "one,two", Spy: ari.DirectionIn, Whisper: ari.DirectionOut})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.URL.EscapedPath() != test.path || len(req.URL.Query()) != 0 {
					t.Errorf("request=%s %s?%s, want POST %s without query", req.Method, req.URL.EscapedPath(), req.URL.RawQuery, test.path)
				}
				wantBody := map[string]any{}
				for name, value := range test.body {
					wantBody[name] = value
				}
				for name, values := range test.query {
					switch name {
					case "media":
						media := make([]any, len(values))
						for i, value := range values {
							media[i] = value
						}
						wantBody[name] = media
					case "offsetms", "skipms":
						n, _ := strconv.Atoi(values[0])
						wantBody[name] = float64(n)
					default:
						wantBody[name] = values[0]
					}
				}
				var body map[string]any
				if req.Body == nil {
					t.Error("missing body")
				} else {
					data, err := io.ReadAll(req.Body)
					if err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal(data, &body); err != nil {
						t.Fatal(err)
					}
					for name, want := range wantBody {
						if !reflect.DeepEqual(body[name], want) {
							t.Errorf("body[%s]=%v want=%v", name, body[name], want)
						}
					}
					if _, present := body[test.absent]; present {
						t.Errorf("unexpected %s in body: %s", test.absent, data)
					}
				}
				return ari23Response(req, http.StatusOK, `{"id":"`+test.responseID+`"}`), nil
			})}})
			id, err := test.call(client)
			if err != nil || id != test.responseID {
				t.Fatalf("kind=%s id=%q want=%q error=%v", test.kind, id, test.responseID, err)
			}
		})
	}
}

func TestServerAssignedRouteRequiresResponseID(t *testing.T) {
	client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
		return ari23Response(req, http.StatusOK, `{}`), nil
	})}})
	if _, err := client.Bridge().CreateWithoutID(nil, ari.BridgeCreateOptions{}); err == nil {
		t.Fatal("missing bridge ID was accepted")
	}
	if _, err := client.Channel().PlayWithoutID(ari.NewKey(ari.ChannelKey, "channel-1"), ari.ChannelPlayOptions{Media: []string{"sound:one"}}); err == nil {
		t.Fatal("missing playback ID was accepted")
	}
}
