package cmd

import (
	"testing"
	"time"

	"github.com/amackera/nornsctl/internal/api"
)

func TestWorkerTools(t *testing.T) {
	cases := []struct {
		name string
		w    api.Worker
		want string
	}{
		{"none", api.Worker{}, "-"},
		{"count only, from an older core", api.Worker{ToolCount: 4}, "4 tools"},
		{"one, count only", api.Worker{ToolCount: 1}, "1 tool"},
		{"a few", api.Worker{ToolCount: 2, Tools: []string{"a", "b"}}, "a,b"},
		{"exactly three", api.Worker{ToolCount: 3, Tools: []string{"a", "b", "c"}}, "a,b,c"},
		{"more are counted", api.Worker{ToolCount: 5, Tools: []string{"a", "b", "c", "d", "e"}}, "a,b,c +2"},
	}
	for _, c := range cases {
		if got := workerTools(c.w); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestConnectedFor(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	ago := func(d time.Duration) *time.Time {
		at := now.Add(-d)
		return &at
	}

	cases := []struct {
		at   *time.Time
		want string
	}{
		{nil, "-"},
		{ago(5 * time.Second), "5s"},
		{ago(3 * time.Minute), "3m"},
		{ago(2 * time.Hour), "2h"},
		{ago(50 * time.Hour), "2d"},
	}
	for _, c := range cases {
		if got := connectedFor(c.at, now); got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
}
