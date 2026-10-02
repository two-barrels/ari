package native

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestBridgeCollectionVariablesUseLiveHandlerShape(t *testing.T) {
	client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/ari/bridges" {
			t.Errorf("request=%s %s", req.Method, req.URL.Path)
		}
		query := req.URL.Query()
		if query.Get("type") != "mixing" || query.Get("name") != "test bridge" || query.Get("bridgeId") != "bridge-1" {
			t.Errorf("query=%v", query)
		}
		var body map[string]json.RawMessage
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body) != 1 || string(body["ARI_TEST_MARKER"]) != `{"value":"created"}` {
			t.Errorf("body=%v, want a flat variable map", body)
		}
		return ari23Response(req, http.StatusOK, `{"id":"bridge-1"}`), nil
	})}})
	bridge, err := client.Bridge().CreateOnCollection(nil, "bridge-1", ari.BridgeCreateOptions{
		Type: "mixing", Name: "test bridge", Variables: map[string]ari.BridgeCreateVariable{
			"ARI_TEST_MARKER": {Value: "created"},
		},
	})
	if err != nil || bridge == nil || bridge.ID() != "bridge-1" {
		t.Fatalf("bridge=%v, error=%v", bridge, err)
	}
}
