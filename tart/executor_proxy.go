package tart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type execPayload struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data,omitempty"`
}

type execResponse struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

// forwardToExecutor sends an action/payload to the Tart executor daemon.
// Target URL can be configured via EXECUTOR_URL env var (default: http://localhost:9090).
func forwardToExecutor(action string, payload interface{}) error {
	cmd := execPayload{Action: action, Data: payload}
	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	base := os.Getenv("EXECUTOR_URL")
	if base == "" {
		base = "http://localhost:9090"
	}
	url := base + "/execute"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("forwardToExecutor(%s) request failed: %v", action, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("executor unexpected status: %s", resp.Status)
	}
	var er execResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return err
	}
	if er.Error != "" {
		return fmt.Errorf("executor error: %s", er.Error)
	}
	return nil
}
