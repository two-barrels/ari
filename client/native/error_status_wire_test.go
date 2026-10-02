package native

import (
	"net/http"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestDataErrorRetainsHTTPStatus(t *testing.T) {
	client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
		return ari23Response(req, http.StatusNotFound, `{"message":"Bridge not found"}`), nil
	})}})
	_, err := client.Bridge().Data(ari.NewKey(ari.BridgeKey, "missing"))
	if CodeFromError(err) != http.StatusNotFound {
		t.Fatalf("error=%v, status=%d, want 404", err, CodeFromError(err))
	}
}
