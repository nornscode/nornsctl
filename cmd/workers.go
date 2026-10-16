package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/amackera/nornsctl/internal/client"
	"github.com/spf13/cobra"
)

var workersCmd = &cobra.Command{
	Use:     "workers",
	Aliases: []string{"worker"},
	Short:   "Inspect connected and declared workers",
}

var workersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workers connected to Norns, and serverless workers asleep",
	Long: `Lists every worker holding a connection right now, with its capabilities,
the tools it registered, the gard it claimed, and whether it is draining.
Below them, every declared serverless worker that is not connected: asleep
until a call to one of its tools wakes it.

The connected list is live, not stored. A worker that is running but missing
here never reached Norns: it is crash-looping or cannot connect.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := &api.WorkerService{Client: newClient()}
		workers, err := svc.List()
		if err != nil {
			return err
		}

		asleep, err := asleepWorkers(svc)
		if err != nil {
			return err
		}

		if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(struct {
				Connected []api.Worker           `json:"connected"`
				Asleep    []api.ServerlessWorker `json:"asleep"`
			}{workers, asleep})
		}

		if len(workers) == 0 && len(asleep) == 0 {
			fmt.Println("No workers connected or declared.")
			return nil
		}

		now := time.Now()
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "WORKER\tSTATUS\tCAPABILITIES\tTOOLS\tGARD\tCONNECTED")
		for _, wk := range workers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				wk.WorkerID,
				workerStatus(wk),
				strings.Join(wk.Capabilities, ","),
				workerTools(wk),
				intOrDash(wk.Gard),
				connectedFor(wk.ConnectedAt, now),
			)
		}
		for _, sw := range asleep {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				sw.WorkerID,
				"asleep",
				strings.Join(sw.Capabilities, ","),
				toolSummary(sw.ToolNames(), len(sw.Tools)),
				"-",
				lastWoken(sw.LastWokenAt, now),
			)
		}
		return w.Flush()
	},
}

// asleepWorkers is the declared serverless workers with no connection. An
// awake one is already in the connected list, flagged serverless; a disabled
// one has withdrawn its tools and is not waiting for anything. A core that
// predates serverless workers has no endpoint for them, which is not an error.
func asleepWorkers(svc *api.WorkerService) ([]api.ServerlessWorker, error) {
	asleep := []api.ServerlessWorker{}

	declared, err := svc.ListServerless()
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
			return asleep, nil
		}
		return nil, err
	}

	for _, sw := range declared {
		if sw.Enabled && !sw.Awake {
			asleep = append(asleep, sw)
		}
	}
	return asleep, nil
}

func workerStatus(w api.Worker) string {
	if w.Draining {
		return "draining"
	}
	return "ready"
}

func workerTools(w api.Worker) string {
	return toolSummary(w.Tools, w.ToolCount)
}

// toolSummary names up to three tools and counts the rest, so the table
// stays one line per worker however many tools it registered.
func toolSummary(names []string, count int) string {
	if count == 0 {
		return "-"
	}
	if len(names) == 0 {
		// An older core reports the count without the names.
		if count == 1 {
			return "1 tool"
		}
		return fmt.Sprintf("%d tools", count)
	}
	const show = 3
	if len(names) <= show {
		return strings.Join(names, ",")
	}
	return fmt.Sprintf("%s +%d", strings.Join(names[:show], ","), len(names)-show)
}

// connectedFor renders how long a worker has been connected, coarsely: the
// column is for telling "just reconnected" from "up all week", not timing.
func connectedFor(at *time.Time, now time.Time) string {
	if at == nil {
		return "-"
	}
	return coarseDuration(now.Sub(*at))
}

// lastWoken fills the CONNECTED column for a worker that is not connected:
// when Norns last tried to wake it, which is the thing to know when a
// declared worker never shows up.
func lastWoken(at *time.Time, now time.Time) string {
	if at == nil {
		return "never woken"
	}
	return "woke " + coarseDuration(now.Sub(*at)) + " ago"
}

func coarseDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func intOrDash(n *int) string {
	if n == nil {
		return "-"
	}
	return strconv.Itoa(*n)
}

func init() {
	rootCmd.AddCommand(workersCmd)
	workersCmd.AddCommand(workersListCmd)

	workersListCmd.Flags().Bool("json", false, "Output as JSON")
}
