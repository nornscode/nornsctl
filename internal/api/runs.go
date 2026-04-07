package api

import (
	"encoding/json"
	"fmt"

	"github.com/amackera/nornsctl/internal/client"
)

type RunService struct {
	Client *client.Client
}

func (s *RunService) List(limit int) ([]Run, error) {
	path := fmt.Sprintf("/runs?limit=%d", limit)
	data, err := s.Client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []Run `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}

func (s *RunService) ListByAgent(agentID int) ([]Run, error) {
	data, err := s.Client.Get(fmt.Sprintf("/agents/%d/runs", agentID))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []Run `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}

func (s *RunService) Get(id int) (*Run, error) {
	data, err := s.Client.Get(fmt.Sprintf("/runs/%d", id))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Run `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *RunService) Events(id int) ([]RunEvent, error) {
	data, err := s.Client.Get(fmt.Sprintf("/runs/%d/events", id))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []RunEvent `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}

type RetryResponse struct {
	Status string `json:"status"`
	RunID  int    `json:"run_id"`
}

func (s *RunService) Retry(id int) (*RetryResponse, error) {
	data, err := s.Client.Post(fmt.Sprintf("/runs/%d/retry", id), nil)
	if err != nil {
		return nil, err
	}
	var resp RetryResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp, nil
}
