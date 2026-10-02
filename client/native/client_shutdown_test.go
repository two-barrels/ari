package native

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestClientConcurrentShutdown(t *testing.T) {
	disconnected := make(chan struct{})
	mux := http.NewServeMux()
	mux.Handle("/ari/events", websocket.Handler(func(ws *websocket.Conn) {
		defer close(disconnected)
		var message []byte
		_ = websocket.Message.Receive(ws, &message)
	}))
	mux.HandleFunc("/ari/asterisk/info", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"system":{"entity_id":"test"}}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := New(&Options{
		URL: server.URL + "/ari", WebsocketURL: "ws" + strings.TrimPrefix(server.URL, "http") + "/ari/events",
		Application: "ari-testing", Username: "test", Password: "test",
	})
	if err := c.ConnectWithContext(ctx); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if !c.Connected() {
		t.Fatal("client did not connect")
	}
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 100; j++ {
				_ = c.Connected()
				c.Close()
			}
		}()
	}
	workers.Wait()
	select {
	case <-disconnected:
	case <-ctx.Done():
		t.Fatal("websocket did not close")
	}
	if c.Connected() {
		t.Fatal("client remains connected after Close")
	}
}
