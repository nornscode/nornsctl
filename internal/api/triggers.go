package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/amackera/nornsctl/internal/client"
)

type Trigger struct {
	ID              int        `json:"id"`
	AgentID         int        `json:"agent_id"`
	Name            string     `json:"name"`
	Cron            string     `json:"cron"`
	Message         string     `json:"message"`
	ConversationKey *string    `json:"conversation_key"`
	Enabled         bool       `json:"enabled"`
	LastFiredAt     *time.Time `json:"last_fired_at"`
	InsertedAt      time.Time  `json:"inserted_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TriggerCreate struct {
	AgentID         int    `json:"agent_id"`
	Name            string `json:"name"`
	Cron            string `json:"cron"`
	Message         string `json:"message"`
	ConversationKey string `json:"conversation_key,omitempty"`
	Enabled         *bool  `json:"enabled,omitempty"`
}

type TriggerUpdate struct {
	Name            *string `json:"name,omitempty"`
	Cron            *string `json:"cron,omitempty"`
	Message         *string `json:"message,omitempty"`
	ConversationKey *string `json:"conversation_key,omitempty"`
	Enabled         *bool   `json:"enabled,omitempty"`
}

type TriggerService struct {
	Client *client.Client
}

func (s *TriggerService) List(agentID int) ([]Trigger, error) {
	path := "/triggers"
	if agentID > 0 {
		path += "?agent_id=" + url.QueryEscape(fmt.Sprintf("%d", agentID))
	}
	data, err := s.Client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []Trigger `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}

func (s *TriggerService) Get(id int) (*Trigger, error) {
	data, err := s.Client.Get(fmt.Sprintf("/triggers/%d", id))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Trigger `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *TriggerService) Create(input TriggerCreate) (*Trigger, error) {
	data, err := s.Client.Post("/triggers", input)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Trigger `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *TriggerService) Update(id int, input TriggerUpdate) (*Trigger, error) {
	data, err := s.Client.Put(fmt.Sprintf("/triggers/%d", id), input)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Trigger `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *TriggerService) Delete(id int) error {
	_, err := s.Client.Delete(fmt.Sprintf("/triggers/%d", id))
	return err
}

type TriggerFireResponse struct {
	Status string `json:"status"`
	RunID  int    `json:"run_id"`
}

func (s *TriggerService) Fire(id int) (*TriggerFireResponse, error) {
	data, err := s.Client.Post(fmt.Sprintf("/triggers/%d/fire", id), nil)
	if err != nil {
		return nil, err
	}
	var resp TriggerFireResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp, nil
}
