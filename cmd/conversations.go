package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/spf13/cobra"
)

var conversationsCmd = &cobra.Command{
	Use:     "conversations",
	Aliases: []string{"convos"},
	Short:   "Manage conversations",
}

var conversationsListCmd = &cobra.Command{
	Use:   "list <agent_id>",
	Short: "List conversations for an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid agent ID: %s", args[0])
		}

		svc := &api.ConversationService{Client: newClient()}
		convos, err := svc.List(agentID)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tKEY\tMESSAGES\tTOKENS\tUPDATED")
		for _, c := range convos {
			fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%s\n", c.ID, c.Key, c.MessageCount, c.TokenEstimate, c.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
		return w.Flush()
	},
}

var conversationsShowCmd = &cobra.Command{
	Use:   "show <agent_id> <key>",
	Short: "Show conversation details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid agent ID: %s", args[0])
		}

		svc := &api.ConversationService{Client: newClient()}
		c, err := svc.Get(agentID, args[1])
		if err != nil {
			return err
		}

		fmt.Printf("ID:             %d\n", c.ID)
		fmt.Printf("Agent ID:       %d\n", c.AgentID)
		fmt.Printf("Key:            %s\n", c.Key)
		if c.Summary != nil {
			fmt.Printf("Summary:        %s\n", *c.Summary)
		}
		fmt.Printf("Message Count:  %d\n", c.MessageCount)
		fmt.Printf("Token Estimate: %d\n", c.TokenEstimate)
		fmt.Printf("Created:        %s\n", c.InsertedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:        %s\n", c.UpdatedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var conversationsDeleteCmd = &cobra.Command{
	Use:   "delete <agent_id> <key>",
	Short: "Delete a conversation",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid agent ID: %s", args[0])
		}

		svc := &api.ConversationService{Client: newClient()}
		if err := svc.Delete(agentID, args[1]); err != nil {
			return err
		}
		fmt.Println("Conversation deleted.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(conversationsCmd)
	conversationsCmd.AddCommand(conversationsListCmd)
	conversationsCmd.AddCommand(conversationsShowCmd)
	conversationsCmd.AddCommand(conversationsDeleteCmd)
}
