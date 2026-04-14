package dev

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	URL              string    `json:"url"`
	APIKey           string    `json:"api_key"`
	SecretKeyBase    string    `json:"secret_key_base"`
	StartedAt        time.Time `json:"started_at"`
	TelemetryAsked   bool      `json:"telemetry_asked"`
	TelemetryEnabled bool      `json:"telemetry_enabled"`
}

func StateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".nornsctl", "dev"), nil
}

func statePath() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

func LoadState() (*State, error) {
	path, err := statePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func SaveState(s *State) error {
	dir, err := StateDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "state.json")
	return os.WriteFile(path, data, 0600)
}
