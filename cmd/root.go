package cmd

import (
	"fmt"
	"os"

	"github.com/amackera/nornsctl/internal/client"
	"github.com/spf13/cobra"
)

var (
	apiURL string
	apiKey string
)

var rootCmd = &cobra.Command{
	Use:   "nornsctl",
	Short: "CLI for the Norns durable agent runtime",
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "url", "", "Norns API URL (env: NORNS_URL)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Norns API key (env: NORNS_API_KEY)")
}

func newClient() *client.Client {
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

	return client.New(url, key)
}
