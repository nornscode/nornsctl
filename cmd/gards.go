package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/amackera/nornsctl/internal/api"
	"github.com/spf13/cobra"
)

var gardsCmd = &cobra.Command{
	Use:     "gards",
	Aliases: []string{"gard"},
	Short:   "Manage gards (worker execution contexts)",
}

var gardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List gards",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := &api.GardService{Client: newClient()}
		gards, err := svc.List()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tTEMPLATE")
		for _, g := range gards {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", g.ID, strOrDash(g.Name), g.Status, strOrDash(g.Template))
		}
		return w.Flush()
	},
}

var gardsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a gard",
	Long:  "Creates a gard record and prints its claim token. The token is shown exactly once — pass it to the worker that will claim the gard.",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		template, _ := cmd.Flags().GetString("template")

		svc := &api.GardService{Client: newClient()}
		g, err := svc.Create(api.GardCreate{Name: name, Template: template})
		if err != nil {
			return err
		}

		fmt.Printf("Created gard %d (%s), status: %s\n", g.ID, strOrDash(g.Name), g.Status)
		fmt.Printf("Claim token (shown once): %s\n", g.ClaimToken)
		fmt.Printf("\nStart a worker in this gard with:\n")
		fmt.Printf("  norns.run(agent, gard=%d, claim_token=\"%s\")\n", g.ID, g.ClaimToken)
		return nil
	},
}

var gardsInspectCmd = &cobra.Command{
	Use:   "inspect <id>",
	Short: "Show gard details and ports",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid gard ID: %s", args[0])
		}

		svc := &api.GardService{Client: newClient()}
		g, err := svc.Get(id)
		if err != nil {
			return err
		}

		fmt.Printf("ID:       %d\n", g.ID)
		fmt.Printf("Name:     %s\n", strOrDash(g.Name))
		fmt.Printf("Status:   %s\n", g.Status)
		fmt.Printf("Template: %s\n", strOrDash(g.Template))
		fmt.Printf("Created:  %s\n", g.InsertedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:  %s\n", g.UpdatedAt.Format("2006-01-02 15:04:05"))

		if len(g.Ports) > 0 {
			fmt.Println("Ports:")
			for _, p := range g.Ports {
				fmt.Printf("  :%d (%s) → %s\n", p.InternalPort, strOrDash(p.Name), strOrDash(p.URL))
			}
		}
		return nil
	},
}

var gardsPortsCmd = &cobra.Command{
	Use:   "ports <id>",
	Short: "List a gard's registered ports",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid gard ID: %s", args[0])
		}

		svc := &api.GardService{Client: newClient()}
		ports, err := svc.Ports(id)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PORT\tNAME\tPROTOCOL\tURL")
		for _, p := range ports {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", p.InternalPort, strOrDash(p.Name), p.Protocol, strOrDash(p.URL))
		}
		return w.Flush()
	},
}

var gardsDestroyCmd = &cobra.Command{
	Use:   "destroy <id>",
	Short: "Destroy a gard (soft-delete; kicks its worker)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid gard ID: %s", args[0])
		}

		force, _ := cmd.Flags().GetBool("force")

		svc := &api.GardService{Client: newClient()}
		if err := svc.Destroy(id, force); err != nil {
			return err
		}
		fmt.Printf("Destroyed gard %d\n", id)
		return nil
	},
}

func strOrDash(s *string) string {
	if s == nil || *s == "" {
		return "-"
	}
	return *s
}

func init() {
	rootCmd.AddCommand(gardsCmd)
	gardsCmd.AddCommand(gardsListCmd)
	gardsCmd.AddCommand(gardsCreateCmd)
	gardsCmd.AddCommand(gardsInspectCmd)
	gardsCmd.AddCommand(gardsPortsCmd)
	gardsCmd.AddCommand(gardsDestroyCmd)

	gardsCreateCmd.Flags().String("name", "", "Gard name")
	gardsCreateCmd.Flags().String("template", "", "Template hint for the provisioner (informational)")

	gardsDestroyCmd.Flags().Bool("force", false, "Destroy even if the gard has an active run")
}
