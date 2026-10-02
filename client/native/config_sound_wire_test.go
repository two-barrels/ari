package native

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestConfigFieldsAndSoundFiltersWire(t *testing.T) {
	tests := []struct {
		name, method, path string
		query              url.Values
		body               map[string]any
		response           string
		call               func(*Client) error
	}{
		{"config fields", "PUT", "/ari/asterisk/config/dynamic/sorcery/endpoint/alice", url.Values{}, map[string]any{"fields": []any{map[string]any{"attribute": "callerid", "value": "Alice <100>"}}}, `{}`, func(c *Client) error {
			return c.Asterisk().Config().Update(ari.NewKey("config", "sorcery/endpoint/alice"), []ari.ConfigTuple{{Attribute: "callerid", Value: "Alice <100>"}})
		}},
		{"sound filters", "GET", "/ari/sounds", url.Values{"lang": {"en+US"}, "format": {"wav"}}, nil, `[]`, func(c *Client) error {
			_, err := c.Sound().List(map[string]string{"lang": "en+US", "format": "wav"}, nil)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
				if req.Method != test.method || req.URL.EscapedPath() != test.path || !reflect.DeepEqual(req.URL.Query(), test.query) {
					t.Errorf("request=%s %s?%s", req.Method, req.URL.EscapedPath(), req.URL.RawQuery)
				}
				if test.body == nil {
					if req.Body != nil {
						t.Error("unexpected body")
					}
				} else {
					data, err := io.ReadAll(req.Body)
					if err != nil {
						t.Fatal(err)
					}
					var got map[string]any
					if err := json.Unmarshal(data, &got); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(got, test.body) {
						t.Errorf("body=%s want=%v", data, test.body)
					}
				}
				return ari23Response(req, http.StatusOK, test.response), nil
			})}})
			if err := test.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}
