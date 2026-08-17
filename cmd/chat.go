package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/amackera/nornsctl/internal/ws"
	"github.com/spf13/cobra"
)

var agentsChatCmd = &cobra.Command{
	Use:   "chat <id>",
	Short: "Interactive REPL with an agent",
	Long: `A terminal conversation with an agent: type a message, watch its tool
activity stream live, read the reply, repeat. When the agent asks a
question (ask_human), answer it inline. The whole session shares one
conversation, so the agent keeps context across turns.

Exit with Ctrl-D, "exit", or "quit". The run keeps its durable state
either way — you can resume the same conversation later with
--conversation-key.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid agent ID: %s", args[0])
		}

		gardID, _ := cmd.Flags().GetInt("gard")
		convKey, _ := cmd.Flags().GetString("conversation-key")
		if convKey == "" {
			convKey = fmt.Sprintf("chat-%d", time.Now().Unix())
		}

		c := newClient()
		agents := &api.AgentService{Client: c}
		runs := &api.RunService{Client: c}

		agent, err := agents.Get(id)
		if err != nil {
			return err
		}

		fmt.Printf("Chatting with %s (agent %d), conversation %q", agent.Name, id, convKey)
		if gardID != 0 {
			fmt.Printf(", gard %d", gardID)
		}
		fmt.Println(". Ctrl-D to exit.")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for {
			fmt.Print("\n> ")
			if !scanner.Scan() {
				fmt.Println()
				return scanner.Err()
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if line == "exit" || line == "quit" {
				return nil
			}

			resp, err := agents.SendMessage(id, api.SendMessageInput{
				Content:         line,
				ConversationKey: convKey,
				GardID:          gardID,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}

			if err := chatTurn(runs, scanner, id, resp.RunID); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
		}
	},
}

// chatTurn streams one run's activity, answering ask_human inline, and
// prints the final output when the run reaches a terminal state.
func chatTurn(runs *api.RunService, scanner *bufio.Scanner, agentID, runID int) error {
	err := ws.Tail(ws.TailConfig{
		BaseURL: resolvedURL(),
		APIKey:  resolvedKey(),
		AgentID: agentID,
		RunID:   runID,
	}, func(ev ws.Event) {
		switch ev.Type {
		case "tool_call":
			fmt.Printf("  ⚙ %s %s\n", str(ev.Payload["name"]), truncate(compactArgs(ev.Payload["arguments"]), 80))
		case "tool_result":
			fmt.Printf("    → %s\n", truncate(str(ev.Payload["content"]), 100))
		case "waiting_for_user":
			fmt.Printf("\n%s\nreply> ", str(ev.Payload["question"]))
			if !scanner.Scan() {
				fmt.Printf("\n(no answer — run %d stays waiting; answer later with: nornsctl runs reply %d \"...\")\n", runID, runID)
				return
			}
			answer := strings.TrimSpace(scanner.Text())
			if answer == "" {
				fmt.Printf("(empty answer ignored — run %d stays waiting)\n", runID)
				return
			}
			if err := runs.Reply(runID, answer); err != nil {
				fmt.Fprintf(os.Stderr, "error delivering reply: %v\n", err)
			}
		}
	})
	if err != nil {
		return err
	}

	// The stream ended (completed or error) — the REST record is the
	// authoritative place for output and failure detail.
	run, err := runs.Get(runID)
	if err != nil {
		return err
	}
	switch run.Status {
	case "completed":
		if run.Output != nil {
			fmt.Printf("\n%s\n", *run.Output)
		}
	case "failed":
		fmt.Printf("\nRun %d failed", runID)
		if fi := run.FailureInspector; fi != nil && fi.ErrorClass != "" {
			fmt.Printf(" (%s)", fi.ErrorClass)
		}
		fmt.Printf(". Events: nornsctl runs events %d\n", runID)
	case "waiting":
		// The user declined to answer inline; nothing more to stream.
	default:
		fmt.Printf("\nRun %d is %s (check: nornsctl runs show %d)\n", runID, run.Status, runID)
	}
	return nil
}

func compactArgs(v any) string {
	m, ok := v.(map[string]any)
	if !ok || len(m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m))
	for k, val := range m {
		parts = append(parts, fmt.Sprintf("%s=%s", k, truncate(str(val), 40)))
	}
	return strings.Join(parts, " ")
}

func str(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.ReplaceAll(s, "\n", " ")
	}
	return fmt.Sprintf("%v", v)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func init() {
	agentsCmd.AddCommand(agentsChatCmd)
	agentsChatCmd.Flags().Int("gard", 0, "Bind every run in the chat to a gard")
	agentsChatCmd.Flags().String("conversation-key", "", "Conversation key (default: a fresh chat-<timestamp>)")
}
