package native

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/two-barrels/ari/v6"
)

type trackedRecordingBody struct {
	*bytes.Reader
	closed bool
	reads  int
}

func (b *trackedRecordingBody) Read(p []byte) (int, error) {
	b.reads++
	return b.Reader.Read(p)
}
func (b *trackedRecordingBody) Close() error { b.closed = true; return nil }

func TestStoredRecordingFileStreamsBinaryAndMetadata(t *testing.T) {
	payload := []byte{0, 1, 2, 0xff, 0x7f}
	body := &trackedRecordingBody{Reader: bytes.NewReader(payload)}
	client := New(&Options{URL: "http://asterisk.test/ari", Username: "user", Password: "secret", HTTPClient: &http.Client{Timeout: time.Millisecond, Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet || req.URL.EscapedPath() != "/ari/recordings/stored/call+1/file" {
			t.Errorf("request = %s %s", req.Method, req.URL.EscapedPath())
		}
		if user, pass, ok := req.BasicAuth(); !ok || user != "user" || pass != "secret" {
			t.Error("missing basic auth")
		}
		resp := ari23Response(req, http.StatusOK, "")
		resp.Body, resp.ContentLength = body, int64(len(payload))
		resp.Header.Set("Content-Type", "audio/wav")
		return resp, nil
	})}})
	file, err := client.StoredRecording().File(context.Background(), ari.NewKey(ari.StoredRecordingKey, "call+1"))
	if err != nil {
		t.Fatal(err)
	}
	if file.ContentType != "audio/wav" || file.Size != int64(len(payload)) || body.reads != 0 {
		t.Fatalf("metadata=%+v reads=%d", file, body.reads)
	}
	time.Sleep(3 * time.Millisecond)
	got, err := io.ReadAll(file.Body)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("body=%v error=%v", got, err)
	}
	if err := file.Body.Close(); err != nil || !body.closed {
		t.Fatalf("close error=%v closed=%v", err, body.closed)
	}
}

func TestStoredRecordingFileStatus(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound} {
		body := &trackedRecordingBody{Reader: bytes.NewReader(nil)}
		client := New(&Options{URL: "http://asterisk.test/ari", HTTPClient: &http.Client{Transport: ari23Transport(func(req *http.Request) (*http.Response, error) {
			resp := ari23Response(req, status, "")
			resp.Body = body
			return resp, nil
		})}})
		file, err := client.StoredRecording().File(context.Background(), ari.NewKey(ari.StoredRecordingKey, "missing"))
		if file != nil || CodeFromError(err) != status || !body.closed {
			t.Fatalf("status=%d file=%+v error=%v closed=%v", status, file, err, body.closed)
		}
	}
}
