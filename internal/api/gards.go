package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/amackera/nornsctl/internal/client"
)

type GardPort struct {
	ID           int     `json:"id"`
	InternalPort int     `json:"internal_port"`
	URL          *string `json:"url"`
	Name         *string `json:"name"`
	Protocol     string  `json:"protocol"`
}

type Gard struct {
	ID         int            `json:"id"`
	Name       *string        `json:"name"`
	Status     string         `json:"status"`
	Template   *string        `json:"template"`
	Metadata   map[string]any `json:"metadata"`
	ClaimToken string         `json:"claim_token,omitempty"` // present only on create
	Ports      []GardPort     `json:"ports,omitempty"`       // present only on show
	InsertedAt time.Time      `json:"inserted_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type GardCreate struct {
	Name     string `json:"name,omitempty"`
	Template string `json:"template,omitempty"`
}

type GardService struct {
	Client *client.Client
}

func (s *GardService) List() ([]Gard, error) {
	data, err := s.Client.Get("/gards")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []Gard `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}

func (s *GardService) Get(id int) (*Gard, error) {
	data, err := s.Client.Get(fmt.Sprintf("/gards/%d", id))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Gard `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *GardService) Create(input GardCreate) (*Gard, error) {
	data, err := s.Client.Post("/gards", input)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Gard `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp.Data, nil
}

func (s *GardService) Destroy(id int, force bool) error {
	path := fmt.Sprintf("/gards/%d", id)
	if force {
		path += "?force=true"
	}
	_, err := s.Client.Delete(path)
	return err
}

func (s *GardService) Ports(id int) ([]GardPort, error) {
	data, err := s.Client.Get(fmt.Sprintf("/gards/%d/ports", id))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []GardPort `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.Data, nil
}
