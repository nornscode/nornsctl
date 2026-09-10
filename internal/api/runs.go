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

// ForkRequest asks for a new run on a run's history up to Step.
type ForkRequest struct {
	Step         int    `json:"step"`
	Message      string `json:"message,omitempty"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	Model        string `json:"model,omitempty"`
}

type ForkResponse struct {
	Status  string `json:"status"`
	RunID   int    `json:"run_id"`
	AgentID int    `json:"agent_id"`
}

// Fork starts a new run from the history of run id after step, in a fresh
// conversation. A system prompt or model override forks onto a variant agent.
func (s *RunService) Fork(id int, req ForkRequest) (*ForkResponse, error) {
	data, err := s.Client.Post(fmt.Sprintf("/runs/%d/fork", id), req)
	if err != nil {
		return nil, err
	}
	var resp ForkResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp, nil
}

// Reply answers a run parked on an ask_human question.
func (s *RunService) Reply(id int, answer string) error {
	body := map[string]string{"answer": answer}
	if _, err := s.Client.Post(fmt.Sprintf("/runs/%d/reply", id), body); err != nil {
		return err
	}
	return nil
}
