package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/spf13/cobra"
)

var workersCmd = &cobra.Command{
	Use:     "workers",
	Aliases: []string{"worker"},
	Short:   "Inspect connected workers",
}

var workersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workers currently connected to Norns",
	Long: `Lists every worker holding a connection right now, with its capabilities,
the tools it registered, the gard it claimed, and whether it is draining.

The list is live, not stored. A worker that is running but missing here
never reached Norns: it is crash-looping or cannot connect.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := &api.WorkerService{Client: newClient()}
		workers, err := svc.List()
		if err != nil {
			return err
		}

		if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(workers)
		}

		if len(workers) == 0 {
			fmt.Println("No workers connected.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "WORKER\tSTATUS\tCAPABILITIES\tTOOLS\tGARD\tCONNECTED")
		for _, wk := range workers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				wk.WorkerID,
				workerStatus(wk),
				strings.Join(wk.Capabilities, ","),
				workerTools(wk),
				intOrDash(wk.Gard),
				connectedFor(wk.ConnectedAt, time.Now()),
			)
		}
		return w.Flush()
	},
}

func workerStatus(w api.Worker) string {
	if w.Draining {
		return "draining"
	}
	return "ready"
}

// workerTools names up to three tools and counts the rest, so the table
// stays one line per worker however many tools it registered.
func workerTools(w api.Worker) string {
	if w.ToolCount == 0 {
		return "-"
	}
	if len(w.Tools) == 0 {
		// An older core reports the count without the names.
		if w.ToolCount == 1 {
			return "1 tool"
		}
		return fmt.Sprintf("%d tools", w.ToolCount)
	}
	const show = 3
	if len(w.Tools) <= show {
		return strings.Join(w.Tools, ",")
	}
	return fmt.Sprintf("%s +%d", strings.Join(w.Tools[:show], ","), len(w.Tools)-show)
}

// connectedFor renders how long a worker has been connected, coarsely: the
// column is for telling "just reconnected" from "up all week", not timing.
func connectedFor(at *time.Time, now time.Time) string {
	if at == nil {
		return "-"
	}
	d := now.Sub(*at)
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
