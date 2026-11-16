package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amackera/nornsctl/internal/client"
)

// A conversation key is caller-chosen text; it must stay one path segment
// even when it holds a slash or a question mark.
func TestConversationKeyIsPathEscaped(t *testing.T) {
	var gotPath, gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.EscapedPath(), r.URL.RawQuery
		w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()

	svc := &ConversationService{Client: client.New(srv.URL, "k")}

	if _, err := svc.Get(3, "slack/C123?x=1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	want := "/api/v1/agents/3/conversations/slack%2FC123%3Fx=1"
	if gotPath != want || gotQuery != "" {
		t.Errorf("get hit %q?%q, want %q", gotPath, gotQuery, want)
	}

	if err := svc.Delete(3, "slack/C123?x=1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if gotPath != want || gotQuery != "" {
		t.Errorf("delete hit %q?%q, want %q", gotPath, gotQuery, want)
	}
}
