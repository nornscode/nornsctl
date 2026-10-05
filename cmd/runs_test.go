package cmd

import (
	"reflect"
	"testing"

	"github.com/amackera/nornsctl/internal/api"
)

func TestCostLines(t *testing.T) {
	gpt := "gpt-5"
	own := &api.Cost{USD: "0.046800"}

	cases := []struct {
		name string
		run  api.Run
		want []string
	}{
		{"a core that predates cost", api.Run{}, nil},
		{
			"no sub-agents",
			api.Run{Cost: own, TotalCost: &api.Cost{USD: "0.046800"}},
			[]string{"Cost:         $0.0468"},
		},
		{
			"sub-agents add to it",
			api.Run{Cost: own, TotalCost: &api.Cost{USD: "0.112345"}},
			[]string{"Cost:         $0.0468", "              $0.1123 with sub-agents"},
		},
		{
			"tokens with no price are named, not counted as free",
			api.Run{
				Cost: &api.Cost{USD: "0.001500", Unpriced: []api.UnpricedUsage{
					{Model: &gpt, InputTokens: 1000, OutputTokens: 100},
					{InputTokens: 50, OutputTokens: 5},
				}},
			},
			[]string{
				"Cost:         $0.0015 + unpriced",
				"Unpriced:     1000 in / 100 out on gpt-5",
				"Unpriced:     50 in / 5 out on unknown model",
			},
		},
		{
			"unpriced tokens below the run are labelled as the tree's",
			api.Run{
				Cost: own,
				TotalCost: &api.Cost{USD: "0.402700", Unpriced: []api.UnpricedUsage{
					{InputTokens: 50, OutputTokens: 5},
				}},
			},
			[]string{
				"Cost:         $0.0468",
				"              $0.4027 + unpriced with sub-agents",
				"Unpriced:     (with sub-agents) 50 in / 5 out on unknown model",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := costLines(&c.run); !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
