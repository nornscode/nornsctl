package cmd

import (
	"time"

	"github.com/amackera/nornsctl/internal/dev"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run a local Norns dev server (foreground)",
	Long:  "Start a local Norns server and Postgres via Docker. Streams logs. Ctrl-C stops everything.\nState is stored in ~/.nornsctl/dev/.",
	RunE:  runDevCmd(false),
}

var devUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the dev server in the background",
	RunE:  runDevCmd(true),
}

var devDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop the dev server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := dev.CheckDockerAvailable(); err != nil {
			return err
		}
		dev.Down()
		return nil
	},
}

var devStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show dev server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := dev.CheckDockerAvailable(); err != nil {
			return err
		}
		dev.StatusInfo()
		return nil
	},
}

var devLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Tail dev server logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := dev.CheckDockerAvailable(); err != nil {
			return err
		}
		return dev.StreamLogs("nornsctl-dev-norns")
	},
}

var devResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Stop dev server and delete all data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := dev.CheckDockerAvailable(); err != nil {
			return err
		}
		dev.Reset()
		return nil
	},
}

func runDevCmd(background bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := dev.CheckDockerAvailable(); err != nil {
			return err
		}

		// Load or create state
		state, err := dev.LoadState()
		if err != nil {
			return err
		}
		if state == nil {
			secretKey, err := dev.GenerateSecretKeyBase()
			if err != nil {
				return err
			}
			apiKey, err := dev.GenerateAPIKey()
			if err != nil {
				return err
			}
			state = &dev.State{
				URL:           "http://localhost:4000",
				APIKey:        apiKey,
				SecretKeyBase: secretKey,
				StartedAt:     time.Now(),
			}
		} else {
			state.StartedAt = time.Now()
		}

		if err := dev.SaveState(state); err != nil {
			return err
		}

		return dev.Up(state, background, rootCmd.Version)
	}
}

func init() {
	rootCmd.AddCommand(devCmd)
	devCmd.AddCommand(devUpCmd)
	devCmd.AddCommand(devDownCmd)
	devCmd.AddCommand(devStatusCmd)
	devCmd.AddCommand(devLogsCmd)
	devCmd.AddCommand(devResetCmd)
}
