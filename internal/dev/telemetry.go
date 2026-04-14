package dev

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// MaybeFirstRunPing prompts the user for anonymous telemetry on first run.
// The answer is stored in state so it only asks once.
func MaybeFirstRunPing(state *State, version string) {
	if state.TelemetryAsked {
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("  Send anonymous first-run telemetry to help improve Norns? [Y/n] ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	state.TelemetryAsked = true
	state.TelemetryEnabled = answer == "" || answer == "y" || answer == "yes"
	SaveState(state)

	if state.TelemetryEnabled {
		sendPing(version)
	}
}

const telemetryURL = "https://norns.mackeracher.com/api/v1/telemetry/first-run"

func sendPing(version string) {
	body, _ := json.Marshal(map[string]string{
		"source":  "nornsctl",
		"version": version,
	})

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(telemetryURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return // fail silently
	}
	resp.Body.Close()
}
