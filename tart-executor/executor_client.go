package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// CommandPayload defines the structure for sending commands
type CommandPayload struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data,omitempty"`
}

// ExecutorResponse represents what the executor returns
type ExecutorResponse struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

func forwardToExecutorDaemon(payload interface{}, action string) error {
	// Build the request payload
	cmd := CommandPayload{
		Action: action,
		Data:   payload,
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	// Change localhost:9090 to a different host/port as needed
	req, err := http.NewRequest("POST", "http://localhost:9090/execute", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("executor POST: %w", err)
	}
	defer resp.Body.Close()

	var execResp ExecutorResponse
	if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
		return fmt.Errorf("decode executor response: %w", err)
	}
	if execResp.Error != "" {
		return fmt.Errorf("executor error: %s", execResp.Error)
	}
	return nil
}

