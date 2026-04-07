package api

import "time"

type Agent struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Purpose         *string    `json:"purpose"`
	Status          string     `json:"status"`
	SystemPrompt    string     `json:"system_prompt"`
	Model           *string    `json:"model"`
	Mode            string     `json:"mode"`
	ContextStrategy string     `json:"context_strategy"`
	ContextWindow   int        `json:"context_window"`
	MaxSteps        *int       `json:"max_steps"`
	InsertedAt      time.Time  `json:"inserted_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type AgentCreate struct {
	Name         string       `json:"name"`
	Purpose      string       `json:"purpose,omitempty"`
	Status       string       `json:"status"`
	SystemPrompt string       `json:"system_prompt"`
	Model        string       `json:"model,omitempty"`
	MaxSteps     *int         `json:"max_steps,omitempty"`
	ModelConfig  *ModelConfig `json:"model_config,omitempty"`
}

type ModelConfig struct {
	Mode             string `json:"mode,omitempty"`
	CheckpointPolicy string `json:"checkpoint_policy,omitempty"`
	ContextStrategy  string `json:"context_strategy,omitempty"`
	ContextWindow    *int   `json:"context_window,omitempty"`
	OnFailure        string `json:"on_failure,omitempty"`
}

type AgentStatus struct {
	Status          string  `json:"status"`
	AgentID         int     `json:"agent_id"`
	RunID           *int    `json:"run_id"`
	Step            int     `json:"step"`
	ConversationID  *int    `json:"conversation_id"`
	ConversationKey string  `json:"conversation_key"`
	MessageCount    int     `json:"message_count"`
}

type Run struct {
	ID               int              `json:"id"`
	AgentID          int              `json:"agent_id"`
	ConversationID   *int             `json:"conversation_id"`
	Status           string           `json:"status"`
	TriggerType      string           `json:"trigger_type"`
	Input            map[string]any   `json:"input"`
	Output           *string          `json:"output"`
	FailureMetadata  map[string]any   `json:"failure_metadata"`
	FailureInspector *FailureInspector `json:"failure_inspector"`
	InsertedAt       time.Time        `json:"inserted_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type FailureInspector struct {
	ErrorClass     string         `json:"error_class"`
	ErrorCode      string         `json:"error_code"`
	RetryDecision  string         `json:"retry_decision"`
	LastCheckpoint map[string]any `json:"last_checkpoint"`
	LastEvent      map[string]any `json:"last_event"`
}

type RunEvent struct {
	ID         int            `json:"id"`
	Sequence   int            `json:"sequence"`
	EventType  string         `json:"event_type"`
	Payload    map[string]any `json:"payload"`
	Source     string         `json:"source"`
	InsertedAt time.Time      `json:"inserted_at"`
}

type Conversation struct {
	ID            int       `json:"id"`
	AgentID       int       `json:"agent_id"`
	Key           string    `json:"key"`
	Summary       *string   `json:"summary"`
	MessageCount  int       `json:"message_count"`
	TokenEstimate int       `json:"token_estimate"`
	InsertedAt    time.Time `json:"inserted_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
