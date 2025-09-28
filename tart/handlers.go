package tart

import (
    "encoding/json"
    "log"
    "net/http"
    "strings"
    "sync"
)

type vmEntry struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Image  string `json:"image"`
    Status string `json:"status"`
}

// isRegistryRef heuristically determines whether an image string refers to a remote
// registry reference (e.g., ghcr.io/org/image:tag) versus a local Tart image name.
// Heuristic: presence of '/' or ':' suggests a remote reference; scheme prefixes also count.
func isRegistryRef(s string) bool {
    if s == "" {
        return false
    }
    if strings.Contains(s, "/") || strings.Contains(s, ":") {
        return true
    }
    if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
        return true
    }
    return false
}

var (
    vmMu    sync.RWMutex
    vmStore = map[string]vmEntry{}
)

func SetupRouter() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/vms", AuthMiddleware(handleVMs))
    mux.HandleFunc("/api/vms/", AuthMiddleware(handleVMByID))
    mux.HandleFunc("/api/images", AuthMiddleware(handleImages))
    return mux
}

func handleVMs(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST":
        // Validate input
        var payload struct {
            Name  string `json:"name"`
            Image string `json:"image"`
        }
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, "invalid json payload", http.StatusBadRequest)
            return
        }
        if payload.Name == "" {
            http.Error(w, "missing name", http.StatusBadRequest)
            return
        }
        // Path 1: URL image -> download + create-from-image
        if strings.HasPrefix(payload.Image, "http://") || strings.HasPrefix(payload.Image, "https://") {
            dl := map[string]string{
                "url":      payload.Image,
                "destName": payload.Name,
            }
            if err := forwardToExecutor("download_image", dl); err != nil {
                log.Printf("executor download_image failed: %v", err)
                http.Error(w, "executor error", http.StatusBadGateway)
                return
            }
            if err := forwardToExecutor("create_vm", payload); err != nil {
                log.Printf("executor create_vm failed: %v", err)
                http.Error(w, "executor error", http.StatusBadGateway)
                return
            }
        } else {
            // Path 2: If image looks like a remote registry ref, pull then clone.
            // Otherwise treat as a local image name and just clone.
            if isRegistryRef(payload.Image) {
                if err := forwardToExecutor("pull_image", map[string]string{"ref": payload.Image}); err != nil {
                    log.Printf("executor pull_image failed: %v", err)
                    http.Error(w, "executor error", http.StatusBadGateway)
                    return
                }
            }
            if err := forwardToExecutor("clone_vm", map[string]string{"name": payload.Name, "image": payload.Image}); err != nil {
                log.Printf("executor clone_vm failed: %v", err)
                http.Error(w, "executor error", http.StatusBadGateway)
                return
            }
        }
        // Persist in store on success
        ent := vmEntry{ID: payload.Name, Name: payload.Name, Image: payload.Image, Status: "running"}
        vmMu.Lock()
        vmStore[ent.ID] = ent
        vmMu.Unlock()
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{"id": ent.ID, "status": ent.Status})
    case "GET":
        // List VMs
        vmMu.RLock()
        list := make([]vmEntry, 0, len(vmStore))
        for _, v := range vmStore {
            list = append(list, v)
        }
        vmMu.RUnlock()
        json.NewEncoder(w).Encode(list)
    default:
        http.Error(w, "method not allowed", 405)
    }
}

func handleVMByID(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/api/vms/")
    // Handle sub-resource /run
    if strings.HasSuffix(path, "/run") {
        id := strings.TrimSuffix(path, "/run")
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        if id == "" {
            http.Error(w, "missing id", http.StatusBadRequest)
            return
        }
        if err := forwardToExecutor("run_vm", map[string]string{"id": id}); err != nil {
            log.Printf("executor run_vm failed: %v", err)
            http.Error(w, "executor error", http.StatusBadGateway)
            return
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"result": "executed"})
        return
    }
    id := path
    switch r.Method {
    case http.MethodGet:
        vmMu.RLock()
        ent, ok := vmStore[id]
        vmMu.RUnlock()
        if !ok {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        json.NewEncoder(w).Encode(ent)
        return
    case http.MethodDelete:
        // Proxy delete to executor and enforce success
        if err := forwardToExecutor("delete_vm", map[string]string{"id": id}); err != nil {
            log.Printf("executor delete_vm failed: %v", err)
            http.Error(w, "executor error", http.StatusBadGateway)
            return
        }
        vmMu.Lock()
        delete(vmStore, id)
        vmMu.Unlock()
        w.WriteHeader(http.StatusNoContent)
        return
    default:
        http.Error(w, "method not allowed", 405)
    }
}

func handleImages(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    // Return a dummy images list
    images := []map[string]string{
        {"id": "debian-13-arm64", "name": "Debian 13 (ARM64)"},
    }
    json.NewEncoder(w).Encode(images)
}
