package dev

import "testing"

func TestResolveNornsImagePrecedence(t *testing.T) {
	t.Setenv(EnvNornsImage, "env-image:latest")
	state := &State{Image: "state-image:latest"}

	got := ResolveNornsImage("flag-image:latest", state)
	if got != "flag-image:latest" {
		t.Fatalf("expected flag image, got %q", got)
	}

	got = ResolveNornsImage("", state)
	if got != "env-image:latest" {
		t.Fatalf("expected env image, got %q", got)
	}
}

func TestResolveNornsImageFallsBackToStateThenDefault(t *testing.T) {
	t.Setenv(EnvNornsImage, "")

	got := ResolveNornsImage("", &State{Image: "state-image:latest"})
	if got != "state-image:latest" {
		t.Fatalf("expected state image, got %q", got)
	}

	got = ResolveNornsImage("", nil)
	if got != DefaultNornsImage {
		t.Fatalf("expected default image, got %q", got)
	}
}
