package dev

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	NetworkName   = "nornsctl-dev"
	PostgresName  = "nornsctl-dev-postgres"
	NornsName     = "nornsctl-dev-norns"
	VolumeName    = "nornsctl-dev-postgres-data"
	PostgresImage = "postgres:16-alpine"
	NornsImage    = "ghcr.io/amackera/norns:main"
	NornsPort     = "4000"
	PostgresUser  = "norns"
	PostgresPass  = "norns"
	PostgresDB    = "norns_dev"
)

// Up starts the dev server. If background is false, streams logs and stops
// containers on Ctrl-C.
func Up(state *State, background bool, version string) error {
	// Check port 4000 is free
	if err := checkPort(NornsPort); err != nil {
		return err
	}

	// Ensure network and volume
	if !NetworkExists(NetworkName) {
		if err := CreateNetwork(NetworkName); err != nil {
			return fmt.Errorf("creating network: %w", err)
		}
	}
	if !VolumeExists(VolumeName) {
		if err := CreateVolume(VolumeName); err != nil {
			return fmt.Errorf("creating volume: %w", err)
		}
	}

	// Pull images
	for _, img := range []string{PostgresImage, NornsImage} {
		if err := PullImage(img); err != nil {
			return fmt.Errorf("pulling %s: %w", img, err)
		}
	}

	// Start Postgres (no host port — only accessible via Docker network)
	if !IsRunning(PostgresName) {
		RemoveContainer(PostgresName)

		restartPolicy := "no"
		if background {
			restartPolicy = "unless-stopped"
		}

		_, err := dockerRun([]string{
			"--name", PostgresName,
			"--network", NetworkName,
			"--restart", restartPolicy,
			"-v", VolumeName + ":/var/lib/postgresql/data",
			"-e", "POSTGRES_USER=" + PostgresUser,
			"-e", "POSTGRES_PASSWORD=" + PostgresPass,
			"-e", "POSTGRES_DB=" + PostgresDB,
			PostgresImage,
		})
		if err != nil {
			return fmt.Errorf("starting postgres: %w", err)
		}
	}

	// Wait for Postgres
	fmt.Print("Waiting for Postgres...")
	if err := waitForPostgres(); err != nil {
		return err
	}
	fmt.Println(" ready")

	// Start Norns
	if !IsRunning(NornsName) {
		RemoveContainer(NornsName)

		restartPolicy := "no"
		if background {
			restartPolicy = "unless-stopped"
		}

		databaseURL := fmt.Sprintf("ecto://%s:%s@%s/%s", PostgresUser, PostgresPass, PostgresName, PostgresDB)

		_, err := dockerRun([]string{
			"--name", NornsName,
			"--network", NetworkName,
			"--restart", restartPolicy,
			"-p", NornsPort + ":4000",
			"-e", "DATABASE_URL=" + databaseURL,
			"-e", "SECRET_KEY_BASE=" + state.SecretKeyBase,
			"-e", "NORNS_DEFAULT_TENANT_KEY=" + state.APIKey,
			"-e", "PHX_HOST=localhost",
			"-e", "PORT=4000",
			NornsImage,
		})
		if err != nil {
			return fmt.Errorf("starting norns: %w", err)
		}
	}

	// Wait for Norns
	fmt.Print("Waiting for Norns...")
	if err := waitForNorns(); err != nil {
		return err
	}
	fmt.Println(" ready")

	fmt.Printf("\n  Norns is running at http://localhost:%s\n", NornsPort)
	fmt.Printf("  API key: %s\n\n", state.APIKey)

	MaybeFirstRunPing(state, version)

	if background {
		fmt.Println("  Stop with: nornsctl dev down")
		return nil
	}

	// Foreground: stream logs, Ctrl-C stops everything
	fmt.Println("  Ctrl-C to stop")
	fmt.Println()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Println("\nStopping...")
		Down()
		os.Exit(0)
	}()

	return StreamLogs(NornsName)
}

// Down stops and removes the dev containers.
func Down() {
	for _, name := range []string{NornsName, PostgresName} {
		StopContainer(name)
		RemoveContainer(name)
	}
	RemoveNetwork(NetworkName)
	fmt.Println("Stopped.")
}

// Reset stops everything and deletes all data.
func Reset() {
	Down()
	if VolumeExists(VolumeName) {
		if err := RemoveVolume(VolumeName); err != nil {
			fmt.Printf("Warning: could not remove volume: %v\n", err)
		} else {
			fmt.Println("Database volume removed.")
		}
	}
	dir, err := StateDir()
	if err == nil {
		os.Remove(dir + "/state.json")
		fmt.Println("State reset.")
	}
}

// StatusInfo prints the current dev server status.
func StatusInfo() {
	state, err := LoadState()
	if err != nil || state == nil {
		fmt.Println("No dev server configured. Run `nornsctl dev` to start one.")
		return
	}

	fmt.Printf("URL:      %s\n", state.URL)
	fmt.Printf("API Key:  %s\n", state.APIKey)
	fmt.Printf("Started:  %s\n", state.StartedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Postgres: %s\n", runStatus(PostgresName))
	fmt.Printf("Norns:    %s\n", runStatus(NornsName))
}

// --- internal ---

func checkPort(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("port %s is already in use. Stop whatever is using it, or set a different port with --port", port)
	}
	ln.Close()
	return nil
}

func waitForPostgres() error {
	for i := 0; i < 30; i++ {
		out, err := docker("exec", PostgresName, "pg_isready", "-U", PostgresUser)
		if err == nil && len(out) > 0 {
			return nil
		}
		time.Sleep(time.Second)
		fmt.Print(".")
	}
	return fmt.Errorf(" postgres did not become healthy in time")
}

func waitForNorns() error {
	url := fmt.Sprintf("http://localhost:%s/api/v1/agents", NornsPort)
	httpClient := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 60; i++ {
		resp, err := httpClient.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(time.Second)
		fmt.Print(".")
	}
	return fmt.Errorf(" norns did not become ready in time")
}

func runStatus(name string) string {
	if IsRunning(name) {
		return "running"
	}
	return "stopped"
}
