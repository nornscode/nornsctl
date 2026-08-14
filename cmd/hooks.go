package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage inbound webhooks",
}

var hooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := &api.HookService{Client: newClient()}
		hooks, err := svc.List()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tAGENT\tSIGNATURE\tENABLED\tPATH")
		for _, h := range hooks {
			fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%t\t%s\n", h.ID, h.Name, h.AgentID, h.SignatureType, h.Enabled, h.Path)
		}
		return w.Flush()
	},
}

var hooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook",
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, _ := cmd.Flags().GetInt("agent")
		name, _ := cmd.Flags().GetString("name")
		messagePath, _ := cmd.Flags().GetString("message-path")
		convKeyPath, _ := cmd.Flags().GetString("conversation-key-path")
		sigType, _ := cmd.Flags().GetString("signature")
		secret, _ := cmd.Flags().GetString("signing-secret")

		svc := &api.HookService{Client: newClient()}
		h, err := svc.Create(api.HookCreate{
			AgentID:             agentID,
			Name:                name,
			MessagePath:         messagePath,
			ConversationKeyPath: convKeyPath,
			SignatureType:       sigType,
			SigningSecret:       secret,
		})
		if err != nil {
			return err
		}

		base := strings.TrimRight(rootURL(), "/")
		fmt.Printf("Created hook %d (%s) on agent %d\n", h.ID, h.Name, h.AgentID)
		fmt.Printf("Deliver to: %s%s\n", base, h.Path)
		return nil
	},
}

var hooksShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show webhook details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid hook ID: %s", args[0])
		}

		svc := &api.HookService{Client: newClient()}
		h, err := svc.Get(id)
		if err != nil {
			return err
		}

		base := strings.TrimRight(rootURL(), "/")
		fmt.Printf("ID:                    %d\n", h.ID)
		fmt.Printf("Name:                  %s\n", h.Name)
		fmt.Printf("Agent ID:              %d\n", h.AgentID)
		fmt.Printf("URL:                   %s%s\n", base, h.Path)
		fmt.Printf("Signature:             %s\n", h.SignatureType)
		fmt.Printf("Enabled:               %t\n", h.Enabled)
		if h.MessagePath != nil && *h.MessagePath != "" {
			fmt.Printf("Message Path:          %s\n", *h.MessagePath)
		}
		if h.ConversationKeyPath != nil && *h.ConversationKeyPath != "" {
			fmt.Printf("Conversation Key Path: %s\n", *h.ConversationKeyPath)
		}
		fmt.Printf("Created:               %s\n", h.InsertedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var hooksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid hook ID: %s", args[0])
		}

		var input api.HookUpdate
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			input.Name = &v
		}
		if cmd.Flags().Changed("message-path") {
			v, _ := cmd.Flags().GetString("message-path")
			input.MessagePath = &v
		}
		if cmd.Flags().Changed("conversation-key-path") {
			v, _ := cmd.Flags().GetString("conversation-key-path")
			input.ConversationKeyPath = &v
		}
		if cmd.Flags().Changed("signature") {
			v, _ := cmd.Flags().GetString("signature")
			input.SignatureType = &v
		}
		if cmd.Flags().Changed("signing-secret") {
			v, _ := cmd.Flags().GetString("signing-secret")
			input.SigningSecret = &v
		}

		svc := &api.HookService{Client: newClient()}
		h, err := svc.Update(id, input)
		if err != nil {
			return err
		}
		fmt.Printf("Updated hook %d (%s)\n", h.ID, h.Name)
		return nil
	},
}

// The server URL the client would use, for printing full delivery URLs.
func rootURL() string {
	if apiURL != "" {
		return apiURL
	}
	if env := os.Getenv("NORNS_URL"); env != "" {
		return env
	}
	return "http://localhost:4000"
}

func setHookEnabled(idArg string, enabled bool) error {
	id, err := strconv.Atoi(idArg)
	if err != nil {
		return fmt.Errorf("invalid hook ID: %s", idArg)
	}

	svc := &api.HookService{Client: newClient()}
	h, err := svc.Update(id, api.HookUpdate{Enabled: &enabled})
	if err != nil {
		return err
	}
	state := "disabled"
	if h.Enabled {
		state = "enabled"
	}
	fmt.Printf("Hook %d (%s) %s\n", h.ID, h.Name, state)
	return nil
}

var hooksEnableCmd = &cobra.Command{
	Use:   "enable <id>",
	Short: "Enable a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setHookEnabled(args[0], true)
	},
}

var hooksDisableCmd = &cobra.Command{
	Use:   "disable <id>",
	Short: "Disable a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setHookEnabled(args[0], false)
	},
}

var hooksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid hook ID: %s", args[0])
		}

		svc := &api.HookService{Client: newClient()}
		if err := svc.Delete(id); err != nil {
			return err
		}
		fmt.Printf("Deleted hook %d\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(hooksCmd)
	hooksCmd.AddCommand(hooksListCmd)
	hooksCmd.AddCommand(hooksCreateCmd)
	hooksCmd.AddCommand(hooksShowCmd)
	hooksCmd.AddCommand(hooksUpdateCmd)
	hooksCmd.AddCommand(hooksEnableCmd)
	hooksCmd.AddCommand(hooksDisableCmd)
	hooksCmd.AddCommand(hooksDeleteCmd)

	hooksCreateCmd.Flags().Int("agent", 0, "Agent ID (required)")
	hooksCreateCmd.Flags().String("name", "", "Hook name, unique per tenant (required)")
	hooksCreateCmd.Flags().String("message-path", "", "Dot-path into the payload used as the message (default: whole payload as JSON)")
	hooksCreateCmd.Flags().String("conversation-key-path", "", "Dot-path whose value keys the conversation (e.g. From)")
	hooksCreateCmd.Flags().String("signature", "none", "Signature verification: none, github, stripe, slack")
	hooksCreateCmd.Flags().String("signing-secret", "", "Provider signing secret (required unless signature is none)")
	hooksCreateCmd.MarkFlagRequired("agent")
	hooksCreateCmd.MarkFlagRequired("name")

	hooksUpdateCmd.Flags().String("name", "", "Hook name")
	hooksUpdateCmd.Flags().String("message-path", "", "Dot-path into the payload used as the message")
	hooksUpdateCmd.Flags().String("conversation-key-path", "", "Dot-path whose value keys the conversation")
	hooksUpdateCmd.Flags().String("signature", "", "Signature verification: none, github, stripe, slack")
	hooksUpdateCmd.Flags().String("signing-secret", "", "Provider signing secret")
}
