package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/amackera/nornsctl/internal/client"
)

// Worker is a worker currently holding a connection to Norns. The list is
// live, not stored: a worker is here while its socket is open and gone the
// moment it isn't.
type Worker struct {
	WorkerID     string     `json:"worker_id"`
	Capabilities []string   `json:"capabilities"`
	Tools        []string   `json:"tools"`
	ToolCount    int        `json:"tool_count"`
	Gard         *int       `json:"gard"`
	Draining     bool       `json:"draining"`
	Serverless   bool       `json:"serverless"`
	ConnectedAt  *time.Time `json:"connected_at"`
}

// ServerlessWorker is a standing declaration: a worker Norns wakes over HTTP
// when there is work for it. Awake says whether an incarnation is connected
// right now; while it is, it also appears in the connected Worker list.
type ServerlessWorker struct {
	ID           int              `json:"id"`
	WorkerID     string           `json:"worker_id"`
	WakeURL      string           `json:"wake_url"`
	Tools        []ServerlessTool `json:"tools"`
	Capabilities []string         `json:"capabilities"`
	Enabled      bool             `json:"enabled"`
	Awake        bool             `json:"awake"`
	LastWokenAt  *time.Time       `json:"last_woken_at"`
}

type ServerlessTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (w ServerlessWorker) ToolNames() []string {
	names := make([]string, 0, len(w.Tools))
	for _, t := range w.Tools {
		names = append(names, t.Name)
	}
	return names
}

type WorkerService struct {
	Client *client.Client
}

func (s *WorkerService) List() ([]Worker, error) {
	data, err := s.Client.Get("/workers")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []Worker `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}

func (s *WorkerService) ListServerless() ([]ServerlessWorker, error) {
	data, err := s.Client.Get("/serverless_workers")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []ServerlessWorker `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}
