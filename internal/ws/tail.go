package ws

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Event represents a streamed agent event from the Phoenix channel.
type Event struct {
	Type    string
	Payload map[string]any
	Time    time.Time
}

// TailConfig holds the parameters for tailing a run.
type TailConfig struct {
	BaseURL  string
	APIKey   string
	AgentID  int
	RunID    int
	Debug    bool
}

// Tail connects to the Phoenix agent channel and streams events.
// It calls onEvent for each event and stops when the run completes or errors.
func Tail(cfg TailConfig, onEvent func(Event)) error {
	wsURL := buildWSURL(cfg.BaseURL, cfg.APIKey)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("websocket connect failed: %w", err)
	}
	defer conn.Close()

	// Join the agent channel
	topic := fmt.Sprintf("agent:%d", cfg.AgentID)
	joinMsg := []any{nil, "1", topic, "phx_join", map[string]any{"run_id": cfg.RunID}}
	if err := conn.WriteJSON(joinMsg); err != nil {
		return fmt.Errorf("join failed: %w", err)
	}

	// Read join reply
	_, replyData, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("reading join reply: %w", err)
	}

	var reply []json.RawMessage
	if err := json.Unmarshal(replyData, &reply); err != nil {
		return fmt.Errorf("parsing join reply: %w", err)
	}

	if len(reply) >= 5 {
		var event string
		json.Unmarshal(reply[3], &event)
		if event == "phx_reply" {
			var payload struct {
				Status string `json:"status"`
			}
			json.Unmarshal(reply[4], &payload)
			if payload.Status != "ok" {
				return fmt.Errorf("channel join failed: %s", string(reply[4]))
			}
		}
	}

	// Start heartbeat
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		ref := 100
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				ref++
				hb := []any{nil, fmt.Sprintf("%d", ref), "phoenix", "heartbeat", map[string]any{}}
				conn.WriteJSON(hb)
			}
		}
	}()
	defer close(done)

	// Read events
	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			return nil // connection closed
		}

		if cfg.Debug {
			fmt.Fprintf(os.Stderr, "DEBUG ws recv: %s\n", string(msgData))
		}

		var msg []json.RawMessage
		if err := json.Unmarshal(msgData, &msg); err != nil || len(msg) < 5 {
			continue
		}

		var event string
		json.Unmarshal(msg[3], &event)

		// Skip Phoenix internal messages
		if event == "phx_reply" || event == "phx_close" || event == "phx_error" || event == "heartbeat" {
			continue
		}

		var payload map[string]any
		json.Unmarshal(msg[4], &payload)

		// Filter by run_id if present in the payload
		if payloadRunID, ok := payload["run_id"]; ok {
			if rid, ok := payloadRunID.(float64); ok && int(rid) != cfg.RunID {
				continue
			}
		}

		onEvent(Event{
			Type:    event,
			Payload: payload,
			Time:    time.Now(),
		})

		// Terminal events
		if event == "completed" || event == "error" {
			return nil
		}
	}
}

func buildWSURL(baseURL, apiKey string) string {
	// http://host:port -> ws://host:port/socket/websocket?token=...&vsn=2.0.0
	wsBase := strings.Replace(baseURL, "https://", "wss://", 1)
	wsBase = strings.Replace(wsBase, "http://", "ws://", 1)
	wsBase = strings.TrimRight(wsBase, "/")

	params := url.Values{}
	params.Set("token", apiKey)
	params.Set("vsn", "2.0.0")

	return fmt.Sprintf("%s/socket/websocket?%s", wsBase, params.Encode())
}
