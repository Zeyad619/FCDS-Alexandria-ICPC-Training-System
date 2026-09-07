package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/botTEST/sendMessage") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient("TEST", server.Client())
	client.BaseURL = server.URL
	if err := client.SendMessage(context.Background(), 12345, "Trainer assignment updated"); err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
}
