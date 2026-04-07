package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage agents",
}

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := &api.AgentService{Client: newClient()}
		agents, err := svc.List()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tMODE\tMODEL")
		for _, a := range agents {
			model := ""
			if a.Model != nil {
				model = *a.Model
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", a.ID, a.Name, a.Status, a.Mode, model)
		}
		return w.Flush()
	},
}

var agentsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show agent details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid agent ID: %s", args[0])
		}

		svc := &api.AgentService{Client: newClient()}
		a, err := svc.Get(id)
		if err != nil {
			return err
		}

		fmt.Printf("ID:               %d\n", a.ID)
		fmt.Printf("Name:             %s\n", a.Name)
		fmt.Printf("Status:           %s\n", a.Status)
		fmt.Printf("Mode:             %s\n", a.Mode)
		if a.Model != nil {
			fmt.Printf("Model:            %s\n", *a.Model)
		}
		if a.Purpose != nil {
			fmt.Printf("Purpose:          %s\n", *a.Purpose)
		}
		fmt.Printf("Context Strategy: %s\n", a.ContextStrategy)
		fmt.Printf("Context Window:   %d\n", a.ContextWindow)
		if a.MaxSteps != nil {
			fmt.Printf("Max Steps:        %d\n", *a.MaxSteps)
		}
		fmt.Printf("System Prompt:    %s\n", a.SystemPrompt)
		fmt.Printf("Created:          %s\n", a.InsertedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:          %s\n", a.UpdatedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var agentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		systemPrompt, _ := cmd.Flags().GetString("system-prompt")
		status, _ := cmd.Flags().GetString("status")
		purpose, _ := cmd.Flags().GetString("purpose")
		model, _ := cmd.Flags().GetString("model")

		input := api.AgentCreate{
			Name:         name,
			SystemPrompt: systemPrompt,
			Status:       status,
			Purpose:      purpose,
			Model:        model,
		}

		svc := &api.AgentService{Client: newClient()}
		a, err := svc.Create(input)
		if err != nil {
			return err
		}
		fmt.Printf("Created agent %d (%s)\n", a.ID, a.Name)
		return nil
	},
}

var agentsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid agent ID: %s", args[0])
		}

		name, _ := cmd.Flags().GetString("name")
		systemPrompt, _ := cmd.Flags().GetString("system-prompt")
		status, _ := cmd.Flags().GetString("status")
		purpose, _ := cmd.Flags().GetString("purpose")
		model, _ := cmd.Flags().GetString("model")

		input := api.AgentCreate{
			Name:         name,
			SystemPrompt: systemPrompt,
			Status:       status,
			Purpose:      purpose,
			Model:        model,
		}

		svc := &api.AgentService{Client: newClient()}
		a, err := svc.Update(id, input)
		if err != nil {
			return err
		}
		fmt.Printf("Updated agent %d (%s)\n", a.ID, a.Name)
		return nil
	},
}

var agentsStatusCmd = &cobra.Command{
	Use:   "status <id>",
	Short: "Get agent process status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid agent ID: %s", args[0])
		}

		svc := &api.AgentService{Client: newClient()}
		s, err := svc.Status(id)
		if err != nil {
			return err
		}

		fmt.Printf("Status:           %s\n", s.Status)
		fmt.Printf("Agent ID:         %d\n", s.AgentID)
		if s.RunID != nil {
			fmt.Printf("Run ID:           %d\n", *s.RunID)
		}
		fmt.Printf("Step:             %d\n", s.Step)
		if s.ConversationID != nil {
			fmt.Printf("Conversation ID:  %d\n", *s.ConversationID)
		}
		fmt.Printf("Conversation Key: %s\n", s.ConversationKey)
		fmt.Printf("Message Count:    %d\n", s.MessageCount)
		return nil
	},
}

var agentsMessageCmd = &cobra.Command{
	Use:   "message <id>",
	Short: "Send a message to an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid agent ID: %s", args[0])
		}

		content, _ := cmd.Flags().GetString("content")
		convKey, _ := cmd.Flags().GetString("conversation-key")

		input := api.SendMessageInput{
			Content:         content,
			ConversationKey: convKey,
		}

		svc := &api.AgentService{Client: newClient()}
		resp, err := svc.SendMessage(id, input)
		if err != nil {
			return err
		}
		fmt.Printf("Message accepted. Run ID: %d\n", resp.RunID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)
	agentsCmd.AddCommand(agentsListCmd)
	agentsCmd.AddCommand(agentsShowCmd)
	agentsCmd.AddCommand(agentsCreateCmd)
	agentsCmd.AddCommand(agentsUpdateCmd)
	agentsCmd.AddCommand(agentsStatusCmd)
	agentsCmd.AddCommand(agentsMessageCmd)

	agentsCreateCmd.Flags().String("name", "", "Agent name (required)")
	agentsCreateCmd.Flags().String("system-prompt", "", "System prompt (required)")
	agentsCreateCmd.Flags().String("status", "active", "Agent status")
	agentsCreateCmd.Flags().String("purpose", "", "Agent purpose")
	agentsCreateCmd.Flags().String("model", "", "LLM model")
	agentsCreateCmd.MarkFlagRequired("name")
	agentsCreateCmd.MarkFlagRequired("system-prompt")

	agentsUpdateCmd.Flags().String("name", "", "Agent name")
	agentsUpdateCmd.Flags().String("system-prompt", "", "System prompt")
	agentsUpdateCmd.Flags().String("status", "", "Agent status")
	agentsUpdateCmd.Flags().String("purpose", "", "Agent purpose")
	agentsUpdateCmd.Flags().String("model", "", "LLM model")

	agentsMessageCmd.Flags().String("content", "", "Message content (required)")
	agentsMessageCmd.Flags().String("conversation-key", "", "Conversation key for multi-turn")
	agentsMessageCmd.MarkFlagRequired("content")
}
