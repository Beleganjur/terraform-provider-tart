package tart

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"
)

// startFakeExecutor spins up a dummy executor that always returns 200 with a JSON body
// so unit tests don't depend on the real executor binary or Tart.
func startFakeExecutor(t *testing.T) *httptest.Server {
    t.Helper()
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"result":"executed"}`))
    }))
    old := os.Getenv("EXECUTOR_URL")
    os.Setenv("EXECUTOR_URL", srv.URL)
    t.Cleanup(func() {
        srv.Close()
        if old == "" {
            os.Unsetenv("EXECUTOR_URL")
        } else {
            os.Setenv("EXECUTOR_URL", old)
        }
    })
    return srv
}

func TestHandleVMsGet(t *testing.T) {
    _ = startFakeExecutor(t)
    srv := httptest.NewServer(SetupRouter())
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/api/vms")
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
}

func TestHandleVMsPost(t *testing.T) {
    _ = startFakeExecutor(t)
    srv := httptest.NewServer(SetupRouter())
    defer srv.Close()

    body, _ := json.Marshal(map[string]string{"name": "vm1", "image": "debian-13-arm64"})
    resp, err := http.Post(srv.URL+"/api/vms", "application/json", bytes.NewReader(body))
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("expected 201, got %d", resp.StatusCode)
    }
}

func TestHandleVMByIDGet(t *testing.T) {
    _ = startFakeExecutor(t)
    srv := httptest.NewServer(SetupRouter())
    defer srv.Close()

    // Create VM first
    createBody, _ := json.Marshal(map[string]string{"name": "dummy-vm-id", "image": "debian-13-arm64"})
    respCreate, err := http.Post(srv.URL+"/api/vms", "application/json", bytes.NewReader(createBody))
    if err != nil {
        t.Fatalf("create request failed: %v", err)
    }
    respCreate.Body.Close()
    if respCreate.StatusCode != http.StatusCreated {
        t.Fatalf("expected 201 on create, got %d", respCreate.StatusCode)
    }

    resp, err := http.Get(srv.URL + "/api/vms/dummy-vm-id")
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
}

func TestHandleVMByIDDelete(t *testing.T) {
    _ = startFakeExecutor(t)
    srv := httptest.NewServer(SetupRouter())
    defer srv.Close()

    // Create VM first
    createBody, _ := json.Marshal(map[string]string{"name": "dummy-vm-id", "image": "debian-13-arm64"})
    respCreate, err := http.Post(srv.URL+"/api/vms", "application/json", bytes.NewReader(createBody))
    if err != nil {
        t.Fatalf("create request failed: %v", err)
    }
    respCreate.Body.Close()
    if respCreate.StatusCode != http.StatusCreated {
        t.Fatalf("expected 201 on create, got %d", respCreate.StatusCode)
    }

    req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/vms/dummy-vm-id", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusNoContent {
        t.Fatalf("expected 204, got %d", resp.StatusCode)
    }
}

func TestHandleImagesGet(t *testing.T) {
    srv := httptest.NewServer(SetupRouter())
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/api/images")
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
}
