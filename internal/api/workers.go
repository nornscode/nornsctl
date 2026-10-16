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
	ConnectedAt  *time.Time `json:"connected_at"`
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
