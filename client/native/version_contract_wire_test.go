package native

import (
	"net/http"
	"strings"
	"testing"

	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/testfixtures"
)

// The transport simulates version-dependent ARI responses. Live-server
// validation belongs to the integration gate.
func TestVersionedARIResponses(t *testing.T) {
	for _, fixture := range testfixtures.Asterisk20_22_23 {
		if fixture.Path != "/bridges/{bridgeId}/variable" && fixture.Path != "/channels/{channelId}/progress" {
			continue
		}
		t.Run(fixture.Path+"@"+fixture.Version, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				wantPath := strings.Replace(fixture.Path, "{bridgeId}", "bridge-1", 1)
				wantPath = strings.Replace(wantPath, "{channelId}", "channel-1", 1)
				if req.Method != fixture.Method || req.URL.Path != "/ari"+wantPath {
					t.Errorf("request=%s %s, want %s /ari%s", req.Method, req.URL.Path, fixture.Method, wantPath)
				}
				if !fixture.Available {
					return ari23Response(req, http.StatusNotFound, `{"message":"Not Found"}`), nil
				}
				if fixture.Method == http.MethodGet {
					return ari23Response(req, http.StatusOK, `{"value":"enabled"}`), nil
				}
				return ari23Response(req, http.StatusNoContent, ""), nil
			})}})
			var err error
			if fixture.Method == http.MethodGet {
				var value string
				value, err = client.Bridge().GetVariable(ari.NewKey(ari.BridgeKey, "bridge-1"), "TEST")
				if fixture.Available && value != "enabled" {
					t.Errorf("value=%q, want enabled", value)
				}
			} else {
				err = client.Channel().Progress(ari.NewKey(ari.ChannelKey, "channel-1"))
			}
			if fixture.Available && err != nil || !fixture.Available && CodeFromError(err) != http.StatusNotFound {
				t.Errorf("available=%t, error=%v, code=%d", fixture.Available, err, CodeFromError(err))
			}
		})
	}
}
