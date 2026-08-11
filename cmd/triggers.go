package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/spf13/cobra"
)

var triggersCmd = &cobra.Command{
	Use:   "triggers",
	Short: "Manage cron triggers",
}

var triggersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List triggers",
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, _ := cmd.Flags().GetInt("agent")

		svc := &api.TriggerService{Client: newClient()}
		triggers, err := svc.List(agentID)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tAGENT\tCRON\tENABLED\tLAST FIRED")
		for _, t := range triggers {
			lastFired := "never"
			if t.LastFiredAt != nil {
				lastFired = t.LastFiredAt.Format("2006-01-02 15:04")
			}
			fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%t\t%s\n", t.ID, t.Name, t.AgentID, t.Cron, t.Enabled, lastFired)
		}
		return w.Flush()
	},
}

var triggersShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show trigger details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid trigger ID: %s", args[0])
		}

		svc := &api.TriggerService{Client: newClient()}
		t, err := svc.Get(id)
		if err != nil {
			return err
		}

		fmt.Printf("ID:               %d\n", t.ID)
		fmt.Printf("Name:             %s\n", t.Name)
		fmt.Printf("Agent ID:         %d\n", t.AgentID)
		fmt.Printf("Cron:             %s\n", t.Cron)
		fmt.Printf("Enabled:          %t\n", t.Enabled)
		fmt.Printf("Message:          %s\n", t.Message)
		if t.ConversationKey != nil {
			fmt.Printf("Conversation Key: %s\n", *t.ConversationKey)
		}
		if t.LastFiredAt != nil {
			fmt.Printf("Last Fired:       %s\n", t.LastFiredAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("Last Fired:       never\n")
		}
		fmt.Printf("Created:          %s\n", t.InsertedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:          %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var triggersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a trigger",
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, _ := cmd.Flags().GetInt("agent")
		name, _ := cmd.Flags().GetString("name")
		cron, _ := cmd.Flags().GetString("cron")
		message, _ := cmd.Flags().GetString("message")
		convKey, _ := cmd.Flags().GetString("conversation-key")
		disabled, _ := cmd.Flags().GetBool("disabled")

		input := api.TriggerCreate{
			AgentID:         agentID,
			Name:            name,
			Cron:            cron,
			Message:         message,
			ConversationKey: convKey,
		}
		if disabled {
			enabled := false
			input.Enabled = &enabled
		}

		svc := &api.TriggerService{Client: newClient()}
		t, err := svc.Create(input)
		if err != nil {
			return err
		}
		fmt.Printf("Created trigger %d (%s): %s on agent %d\n", t.ID, t.Name, t.Cron, t.AgentID)
		return nil
	},
}

var triggersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid trigger ID: %s", args[0])
		}

		var input api.TriggerUpdate
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			input.Name = &v
		}
		if cmd.Flags().Changed("cron") {
			v, _ := cmd.Flags().GetString("cron")
			input.Cron = &v
		}
		if cmd.Flags().Changed("message") {
			v, _ := cmd.Flags().GetString("message")
			input.Message = &v
		}
		if cmd.Flags().Changed("conversation-key") {
			v, _ := cmd.Flags().GetString("conversation-key")
			input.ConversationKey = &v
		}

		svc := &api.TriggerService{Client: newClient()}
		t, err := svc.Update(id, input)
		if err != nil {
			return err
		}
		fmt.Printf("Updated trigger %d (%s)\n", t.ID, t.Name)
		return nil
	},
}

func setTriggerEnabled(idArg string, enabled bool) error {
	id, err := strconv.Atoi(idArg)
	if err != nil {
		return fmt.Errorf("invalid trigger ID: %s", idArg)
	}

	svc := &api.TriggerService{Client: newClient()}
	t, err := svc.Update(id, api.TriggerUpdate{Enabled: &enabled})
	if err != nil {
		return err
	}
	state := "disabled"
	if t.Enabled {
		state = "enabled"
	}
	fmt.Printf("Trigger %d (%s) %s\n", t.ID, t.Name, state)
	return nil
}

var triggersEnableCmd = &cobra.Command{
	Use:   "enable <id>",
	Short: "Enable a trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setTriggerEnabled(args[0], true)
	},
}

var triggersDisableCmd = &cobra.Command{
	Use:   "disable <id>",
	Short: "Disable a trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setTriggerEnabled(args[0], false)
	},
}

var triggersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid trigger ID: %s", args[0])
		}

		svc := &api.TriggerService{Client: newClient()}
		if err := svc.Delete(id); err != nil {
			return err
		}
		fmt.Printf("Deleted trigger %d\n", id)
		return nil
	},
}

var triggersFireCmd = &cobra.Command{
	Use:   "fire <id>",
	Short: "Fire a trigger now, outside its schedule",
	Long:  "Starts a run immediately with the trigger's message. Does not consume the next scheduled firing.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid trigger ID: %s", args[0])
		}

		svc := &api.TriggerService{Client: newClient()}
		resp, err := svc.Fire(id)
		if err != nil {
			return err
		}
		fmt.Printf("Fired. Run ID: %d\n", resp.RunID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(triggersCmd)
	triggersCmd.AddCommand(triggersListCmd)
	triggersCmd.AddCommand(triggersShowCmd)
	triggersCmd.AddCommand(triggersCreateCmd)
	triggersCmd.AddCommand(triggersUpdateCmd)
	triggersCmd.AddCommand(triggersEnableCmd)
	triggersCmd.AddCommand(triggersDisableCmd)
	triggersCmd.AddCommand(triggersDeleteCmd)
	triggersCmd.AddCommand(triggersFireCmd)

	triggersListCmd.Flags().Int("agent", 0, "Filter by agent ID")

	triggersCreateCmd.Flags().Int("agent", 0, "Agent ID (required)")
	triggersCreateCmd.Flags().String("name", "", "Trigger name, unique per tenant (required)")
	triggersCreateCmd.Flags().String("cron", "", "Cron expression, e.g. '0 9 * * 5' or '@daily' (required)")
	triggersCreateCmd.Flags().String("message", "", "Message sent to the agent on each firing (required)")
	triggersCreateCmd.Flags().String("conversation-key", "", "Persistent conversation key; omit for a fresh conversation per firing")
	triggersCreateCmd.Flags().Bool("disabled", false, "Create the trigger disabled")
	triggersCreateCmd.MarkFlagRequired("agent")
	triggersCreateCmd.MarkFlagRequired("name")
	triggersCreateCmd.MarkFlagRequired("cron")
	triggersCreateCmd.MarkFlagRequired("message")

	triggersUpdateCmd.Flags().String("name", "", "Trigger name")
	triggersUpdateCmd.Flags().String("cron", "", "Cron expression")
	triggersUpdateCmd.Flags().String("message", "", "Message sent to the agent")
	triggersUpdateCmd.Flags().String("conversation-key", "", "Persistent conversation key")
}
