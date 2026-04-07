package api

import (
	"encoding/json"
	"fmt"

	"github.com/amackera/nornsctl/internal/client"
)

type ConversationService struct {
	Client *client.Client
}

func (s *ConversationService) List(agentID int) ([]Conversation, error) {
	data, err := s.Client.Get(fmt.Sprintf("/agents/%d/conversations", agentID))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []Conversation `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}

func (s *ConversationService) Get(agentID int, key string) (*Conversation, error) {
	data, err := s.Client.Get(fmt.Sprintf("/agents/%d/conversations/%s", agentID, key))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Conversation `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *ConversationService) Delete(agentID int, key string) error {
	_, err := s.Client.Delete(fmt.Sprintf("/agents/%d/conversations/%s", agentID, key))
	return err
}
