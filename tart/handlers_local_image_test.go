package tart

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
)

// Verifies that when posting a VM create with a local image name (no '/' or ':'),
// the API skips pull_image and only calls clone_vm on the executor.
func TestHandleVMsPost_LocalImage_SkipsPull(t *testing.T) {
	var pullCount int32
	var cloneCount int32

	mux := http.NewServeMux()
	mux.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct{
			Action string `json:"action"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Action == "pull_image" {
			atomic.AddInt32(&pullCount, 1)
		} else if req.Action == "clone_vm" {
			atomic.AddInt32(&cloneCount, 1)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"executed"}`))
	})
	execSrv := httptest.NewServer(mux)
	defer execSrv.Close()
	old := os.Getenv("EXECUTOR_URL")
	os.Setenv("EXECUTOR_URL", execSrv.URL)
	defer func(){ if old==""{os.Unsetenv("EXECUTOR_URL")}else{os.Setenv("EXECUTOR_URL",old)} }()

	srv := httptest.NewServer(SetupRouter())
	defer srv.Close()

	body, _ := json.Marshal(map[string]string{"name":"vm-local","image":"localimage"})
	resp, err := http.Post(srv.URL+"/api/vms", "application/json", bytes.NewReader(body))
	if err != nil { t.Fatalf("request failed: %v", err) }
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	if atomic.LoadInt32(&pullCount) != 0 {
		t.Fatalf("expected no pull_image calls for local image, got %d", pullCount)
	}
	if atomic.LoadInt32(&cloneCount) != 1 {
		t.Fatalf("expected exactly one clone_vm call, got %d", cloneCount)
	}
}
