package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amackera/nornsctl/internal/client"
)

func TestArchiveCallsDeleteAndPassesForce(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	svc := &AgentService{Client: client.New(srv.URL, "k")}

	if err := svc.Archive(7, false); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/v1/agents/7" {
		t.Errorf("got %s %s, want DELETE /api/v1/agents/7", gotMethod, gotPath)
	}
	if gotQuery != "" {
		t.Errorf("unforced archive sent %q, want no query", gotQuery)
	}

	if err := svc.Archive(7, true); err != nil {
		t.Fatalf("forced archive: %v", err)
	}
	if gotQuery != "force=true" {
		t.Errorf("forced archive sent %q, want force=true", gotQuery)
	}
}

// A refusal has to reach the user as the reason, not as a bare status code:
// "has an active run" is the whole point of the 409.
func TestArchiveSurfacesTheRefusal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":"agent has an active run — pass force=true to archive anyway"}`))
	}))
	defer srv.Close()

	svc := &AgentService{Client: client.New(srv.URL, "k")}

	err := svc.Archive(7, false)
	if err == nil {
		t.Fatal("archive of a busy agent returned no error")
	}
	if !strings.Contains(err.Error(), "active run") {
		t.Errorf("error %q does not say why it was refused", err)
	}
}
