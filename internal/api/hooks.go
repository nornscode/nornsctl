package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/amackera/nornsctl/internal/client"
)

type Hook struct {
	ID                  int       `json:"id"`
	AgentID             int       `json:"agent_id"`
	Name                string    `json:"name"`
	Token               string    `json:"token"`
	Path                string    `json:"path"`
	MessagePath         *string   `json:"message_path"`
	ConversationKeyPath *string   `json:"conversation_key_path"`
	SignatureType       string    `json:"signature_type"`
	Enabled             bool      `json:"enabled"`
	InsertedAt          time.Time `json:"inserted_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type HookCreate struct {
	AgentID             int    `json:"agent_id"`
	Name                string `json:"name"`
	MessagePath         string `json:"message_path,omitempty"`
	ConversationKeyPath string `json:"conversation_key_path,omitempty"`
	SignatureType       string `json:"signature_type,omitempty"`
	SigningSecret       string `json:"signing_secret,omitempty"`
}

type HookUpdate struct {
	Name                *string `json:"name,omitempty"`
	MessagePath         *string `json:"message_path,omitempty"`
	ConversationKeyPath *string `json:"conversation_key_path,omitempty"`
	SignatureType       *string `json:"signature_type,omitempty"`
	SigningSecret       *string `json:"signing_secret,omitempty"`
	Enabled             *bool   `json:"enabled,omitempty"`
}

type HookService struct {
	Client *client.Client
}

func (s *HookService) List() ([]Hook, error) {
	data, err := s.Client.Get("/hooks")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []Hook `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}

func (s *HookService) Get(id int) (*Hook, error) {
	data, err := s.Client.Get(fmt.Sprintf("/hooks/%d", id))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Hook `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *HookService) Create(input HookCreate) (*Hook, error) {
	data, err := s.Client.Post("/hooks", input)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Hook `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *HookService) Update(id int, input HookUpdate) (*Hook, error) {
	data, err := s.Client.Put(fmt.Sprintf("/hooks/%d", id), input)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Hook `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *HookService) Delete(id int) error {
	_, err := s.Client.Delete(fmt.Sprintf("/hooks/%d", id))
	return err
}
