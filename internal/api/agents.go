package api

import (
	"encoding/json"
	"fmt"

	"github.com/amackera/nornsctl/internal/client"
)

type AgentService struct {
	Client *client.Client
}

func (s *AgentService) List() ([]Agent, error) {
	data, err := s.Client.Get("/agents")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []Agent `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}

func (s *AgentService) Get(id int) (*Agent, error) {
	data, err := s.Client.Get(fmt.Sprintf("/agents/%d", id))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Agent `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *AgentService) Create(input AgentCreate) (*Agent, error) {
	data, err := s.Client.Post("/agents", input)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Agent `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *AgentService) Update(id int, input AgentCreate) (*Agent, error) {
	data, err := s.Client.Put(fmt.Sprintf("/agents/%d", id), input)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Agent `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *AgentService) Status(agentID int) (*AgentStatus, error) {
	data, err := s.Client.Get(fmt.Sprintf("/agents/%d/status", agentID))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data AgentStatus `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

type SendMessageInput struct {
	Content         string `json:"content"`
	ConversationKey string `json:"conversation_key,omitempty"`
}

type SendMessageResponse struct {
	Status string `json:"status"`
	RunID  int    `json:"run_id"`
}

func (s *AgentService) SendMessage(agentID int, input SendMessageInput) (*SendMessageResponse, error) {
	data, err := s.Client.Post(fmt.Sprintf("/agents/%d/messages", agentID), input)
	if err != nil {
		return nil, err
	}
	var resp SendMessageResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp, nil
}
