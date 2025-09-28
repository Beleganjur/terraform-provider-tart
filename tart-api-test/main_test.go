package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "github.com/beleganjur/terraform-provider-tart/tart"
    "github.com/stretchr/testify/assert"
)

var (
    apiURL   string
    apiToken = getenv("TART_API_TOKEN", "")
)

func getenv(key, fallback string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return fallback
}

// executorStub starts a local HTTP server that simulates the executor /execute endpoint.
// It always returns 200 OK with {"result":"executed"}.
func executorStub(t *testing.T) *httptest.Server {
    t.Helper()
    mux := http.NewServeMux()
    mux.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"result":"executed"}`))
    })
    return httptest.NewServer(mux)
}
func TestListVMs(t *testing.T) {
    execSrv := executorStub(t)
    defer execSrv.Close()
    os.Setenv("EXECUTOR_URL", execSrv.URL)
    defer os.Unsetenv("EXECUTOR_URL")
    
    srv := httptest.NewServer(tart.SetupRouter())
    defer srv.Close()
    base := srv.URL + "/api"
    
    // Create
    body, _ := json.Marshal(map[string]string{"name": "list-test", "image": "debian-13"})
    req, _ := http.NewRequest(http.MethodPost, base+"/vms", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    _, _ = http.DefaultClient.Do(req)
    // Now list
    req, _ = http.NewRequest(http.MethodGet, base+"/vms", nil)
    resp, _ := http.DefaultClient.Do(req)
    var vms []map[string]interface{}
    _ = json.NewDecoder(resp.Body).Decode(&vms)
    t.Logf("VMs: %+v", vms)
}

func TestListImages(t *testing.T) {
    srv := httptest.NewServer(tart.SetupRouter())
    defer srv.Close()

    req, _ := http.NewRequest(http.MethodGet, apiURL+"/images", nil)
    if apiToken != "" {
        req.Header.Set("Authorization", "Bearer "+apiToken)
    }
    resp, err := http.DefaultClient.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    defer resp.Body.Close()
    var images []map[string]interface{}
    _ = json.NewDecoder(resp.Body).Decode(&images)
    t.Logf("Images: %+v", images)
}

// Test a full VM lifecycle using the API with a stubbed executor
func TestVMLifecycle_CreateGetDelete(t *testing.T) {
    execSrv := executorStub(t)
    defer execSrv.Close()
    // Point API to the stubbed executor
    os.Setenv("EXECUTOR_URL", execSrv.URL)
    defer os.Unsetenv("EXECUTOR_URL")

    apiSrv := httptest.NewServer(tart.SetupRouter())
    defer apiSrv.Close()
    base := apiSrv.URL + "/api"

    // Create VM
    payload := map[string]string{"name": "test-trash", "image": "debian-13"}
    b, _ := json.Marshal(payload)
    req, _ := http.NewRequest(http.MethodPost, base+"/vms", bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    if apiToken != "" { req.Header.Set("Authorization", "Bearer "+apiToken) }
    resp, err := http.DefaultClient.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    resp.Body.Close()

    // GET by ID
    req, _ = http.NewRequest(http.MethodGet, base+"/vms/test-trash", nil)
    if apiToken != "" { req.Header.Set("Authorization", "Bearer "+apiToken) }
    resp, err = http.DefaultClient.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    var got map[string]interface{}
    _ = json.NewDecoder(resp.Body).Decode(&got)
    resp.Body.Close()
    assert.Equal(t, "test-trash", got["name"])

    // List should include it
    req, _ = http.NewRequest(http.MethodGet, base+"/vms", nil)
    if apiToken != "" { req.Header.Set("Authorization", "Bearer "+apiToken) }
    resp, err = http.DefaultClient.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    var list []map[string]interface{}
    _ = json.NewDecoder(resp.Body).Decode(&list)
    resp.Body.Close()
    found := false
    for _, it := range list { if it["name"] == "test-trash" { found = true; break } }
    assert.True(t, found, "expected test-trash in list")

    // Delete VM
    req, _ = http.NewRequest(http.MethodDelete, base+"/vms/test-trash", nil)
    if apiToken != "" { req.Header.Set("Authorization", "Bearer "+apiToken) }
    resp, err = http.DefaultClient.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusNoContent, resp.StatusCode)
    resp.Body.Close()

    // GET should now be 404
    req, _ = http.NewRequest(http.MethodGet, base+"/vms/test-trash", nil)
    if apiToken != "" { req.Header.Set("Authorization", "Bearer "+apiToken) }
    resp, err = http.DefaultClient.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

