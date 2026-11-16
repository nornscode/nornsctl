package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amackera/nornsctl/internal/dev"
)

// The .env wired from dev state holds the API key, so it must not be
// readable by other users.
func TestDevWiredEnvIsPrivate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := dev.SaveState(&dev.State{URL: "http://localhost:4000", APIKey: "nrn_secret"}); err != nil {
		t.Fatalf("saving dev state: %v", err)
	}

	out := filepath.Join(t.TempDir(), "my-agent")
	if err := Run(Config{Name: "my-agent", PackageName: "my_agent", Language: "python", Template: "default", OutputDir: out}); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	info, err := os.Stat(filepath.Join(out, ".env"))
	if err != nil {
		t.Fatalf("stat .env: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf(".env mode is %o, want 600", perm)
	}
}
