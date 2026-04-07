package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/spf13/cobra"
)

var runsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Manage runs",
}

var runsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List runs",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := &api.RunService{Client: newClient()}
		agentID, _ := cmd.Flags().GetInt("agent")
		limit, _ := cmd.Flags().GetInt("limit")

		var runs []api.Run
		var err error
		if agentID > 0 {
			runs, err = svc.ListByAgent(agentID)
		} else {
			runs, err = svc.List(limit)
		}
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tAGENT\tSTATUS\tTRIGGER\tCREATED")
		for _, r := range runs {
			fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%s\n", r.ID, r.AgentID, r.Status, r.TriggerType, r.InsertedAt.Format("2006-01-02 15:04:05"))
		}
		return w.Flush()
	},
}

var runsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show run details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid run ID: %s", args[0])
		}

		svc := &api.RunService{Client: newClient()}
		r, err := svc.Get(id)
		if err != nil {
			return err
		}

		fmt.Printf("ID:           %d\n", r.ID)
		fmt.Printf("Agent ID:     %d\n", r.AgentID)
		fmt.Printf("Status:       %s\n", r.Status)
		fmt.Printf("Trigger:      %s\n", r.TriggerType)
		if r.ConversationID != nil {
			fmt.Printf("Conversation: %d\n", *r.ConversationID)
		}
		if r.Output != nil {
			fmt.Printf("Output:       %s\n", *r.Output)
		}
		fmt.Printf("Created:      %s\n", r.InsertedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:      %s\n", r.UpdatedAt.Format("2006-01-02 15:04:05"))

		if r.FailureInspector != nil {
			fmt.Println("\nFailure Inspector:")
			fmt.Printf("  Error Class:    %s\n", r.FailureInspector.ErrorClass)
			fmt.Printf("  Error Code:     %s\n", r.FailureInspector.ErrorCode)
			fmt.Printf("  Retry Decision: %s\n", r.FailureInspector.RetryDecision)
		}

		return nil
	},
}

var runsEventsCmd = &cobra.Command{
	Use:   "events <id>",
	Short: "Print event log for a run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid run ID: %s", args[0])
		}

		svc := &api.RunService{Client: newClient()}
		events, err := svc.Events(id)
		if err != nil {
			return err
		}

		jsonFlag, _ := cmd.Flags().GetBool("json")
		if jsonFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(events)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SEQ\tTYPE\tSOURCE\tTIME")
		for _, e := range events {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", e.Sequence, e.EventType, e.Source, e.InsertedAt.Format("15:04:05"))
		}
		return w.Flush()
	},
}

var runsRetryCmd = &cobra.Command{
	Use:   "retry <id>",
	Short: "Retry a failed run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid run ID: %s", args[0])
		}

		svc := &api.RunService{Client: newClient()}
		resp, err := svc.Retry(id)
		if err != nil {
			return err
		}
		fmt.Printf("Retry accepted. New run ID: %d\n", resp.RunID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runsCmd)
	runsCmd.AddCommand(runsListCmd)
	runsCmd.AddCommand(runsShowCmd)
	runsCmd.AddCommand(runsEventsCmd)
	runsCmd.AddCommand(runsRetryCmd)

	runsListCmd.Flags().Int("agent", 0, "Filter by agent ID")
	runsListCmd.Flags().Int("limit", 50, "Max number of runs to return")

	runsEventsCmd.Flags().Bool("json", false, "Output events as JSON")
}
