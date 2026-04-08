package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/amackera/nornsctl/internal/ws"
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

var runsTailCmd = &cobra.Command{
	Use:   "tail <id>",
	Short: "Stream events for a run in real-time",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid run ID: %s", args[0])
		}

		c := newClient()
		svc := &api.RunService{Client: c}

		// Get the run to find agent_id
		run, err := svc.Get(id)
		if err != nil {
			return err
		}

		// Print existing events first
		events, err := svc.Events(id)
		if err != nil {
			return err
		}

		for _, e := range events {
			printTailEvent(e.InsertedAt.Format("15:04:05"), e.EventType, eventSummary(e))
		}

		// If run is already terminal, we're done
		if run.Status == "completed" || run.Status == "failed" {
			return nil
		}

		// Stream new events via WebSocket
		url := apiURL
		if url == "" {
			url = os.Getenv("NORNS_URL")
		}
		if url == "" {
			url = "http://localhost:4000"
		}

		key := apiKey
		if key == "" {
			key = os.Getenv("NORNS_API_KEY")
		}

		return ws.Tail(ws.TailConfig{
			BaseURL: url,
			APIKey:  key,
			AgentID: run.AgentID,
			RunID:   id,
		}, func(e ws.Event) {
			printTailEvent(e.Time.Format("15:04:05"), e.Type, wsSummary(e))
		})
	},
}

func printTailEvent(ts, eventType, summary string) {
	color := eventColor(eventType)
	fmt.Printf("%s %s%-16s\033[0m %s\n", ts, color, eventType, summary)
}

func eventColor(eventType string) string {
	switch eventType {
	case "llm_request", "llm_response":
		return "\033[34m" // blue
	case "tool_call":
		return "\033[33m" // yellow
	case "tool_result":
		return "\033[33m" // yellow
	case "run_completed", "completed":
		return "\033[32m" // green
	case "run_failed", "error":
		return "\033[31m" // red
	case "waiting_for_timer":
		return "\033[33m" // yellow
	case "checkpoint_saved":
		return "\033[90m" // gray
	default:
		return "\033[90m" // gray
	}
}

func eventSummary(e api.RunEvent) string {
	p := e.Payload
	switch e.EventType {
	case "llm_response":
		fr, _ := p["finish_reason"].(string)
		return fr
	case "tool_call":
		name, _ := p["name"].(string)
		return name
	case "tool_result":
		name, _ := p["name"].(string)
		content, _ := p["content"].(string)
		if len(content) > 60 {
			content = content[:60] + "..."
		}
		return fmt.Sprintf("%s → %s", name, content)
	case "checkpoint_saved":
		step, _ := p["step"].(float64)
		return fmt.Sprintf("step %d", int(step))
	case "run_completed":
		output, _ := p["output"].(string)
		if len(output) > 80 {
			output = output[:80] + "..."
		}
		return output
	case "run_failed":
		errMsg, _ := p["error"].(string)
		if len(errMsg) > 80 {
			errMsg = errMsg[:80] + "..."
		}
		return errMsg
	case "waiting_for_timer":
		secs, _ := p["seconds"].(float64)
		return fmt.Sprintf("%ds", int(secs))
	case "llm_request":
		step, _ := p["step"].(float64)
		count, _ := p["message_count"].(float64)
		return fmt.Sprintf("step %d, %d messages", int(step), int(count))
	default:
		return ""
	}
}

func wsSummary(e ws.Event) string {
	p := e.Payload
	switch e.Type {
	case "tool_call":
		name, _ := p["name"].(string)
		return name
	case "tool_result":
		name, _ := p["name"].(string)
		content, _ := p["content"].(string)
		if len(content) > 60 {
			content = content[:60] + "..."
		}
		return fmt.Sprintf("%s → %s", name, content)
	case "completed":
		output, _ := p["output"].(string)
		if len(output) > 80 {
			output = output[:80] + "..."
		}
		return output
	case "error":
		errMsg, _ := p["error"].(string)
		return errMsg
	case "waiting_timer":
		secs, _ := p["seconds"].(float64)
		return fmt.Sprintf("%ds", int(secs))
	default:
		return ""
	}
}

func init() {
	rootCmd.AddCommand(runsCmd)
	runsCmd.AddCommand(runsListCmd)
	runsCmd.AddCommand(runsShowCmd)
	runsCmd.AddCommand(runsEventsCmd)
	runsCmd.AddCommand(runsRetryCmd)
	runsCmd.AddCommand(runsTailCmd)

	runsListCmd.Flags().Int("agent", 0, "Filter by agent ID")
	runsListCmd.Flags().Int("limit", 50, "Max number of runs to return")

	runsEventsCmd.Flags().Bool("json", false, "Output events as JSON")
}
