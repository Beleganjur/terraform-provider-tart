package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

func handleExecute(w http.ResponseWriter, r *http.Request) {
    type cmdRequest struct {
        Action string `json:"action"`
        Data   json.RawMessage `json:"data"`
    }
    w.Header().Set("Content-Type", "application/json")
    var req cmdRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, fmt.Sprintf("invalid json: %v", err), http.StatusBadRequest)
        return
    }

    switch req.Action {
    case "pull_image":
        // Expect { ref: string }
        var payload struct{ Ref string `json:"ref"` }
        if err := json.Unmarshal(req.Data, &payload); err != nil {
            http.Error(w, fmt.Sprintf("invalid data: %v", err), http.StatusBadRequest)
            return
        }
        if payload.Ref == "" {
            http.Error(w, "missing ref", http.StatusBadRequest)
            return
        }
        if err := execTart("pull", payload.Ref); err != nil {
            http.Error(w, fmt.Sprintf("tart pull failed: %v", err), http.StatusBadGateway)
            return
        }
        json.NewEncoder(w).Encode(map[string]string{"result": "executed"})
        return

    case "clone_vm":
        // Expect { name: string, image: string }
        var payload struct{ Name, Image string }
        if err := json.Unmarshal(req.Data, &payload); err != nil {
            http.Error(w, fmt.Sprintf("invalid data: %v", err), http.StatusBadRequest)
            return
        }
        if payload.Name == "" || payload.Image == "" {
            http.Error(w, "missing name or image", http.StatusBadRequest)
            return
        }
        if err := execTart("clone", payload.Image, payload.Name); err != nil {
            http.Error(w, fmt.Sprintf("tart clone failed: %v", err), http.StatusBadGateway)
            return
        }
        json.NewEncoder(w).Encode(map[string]string{"result": "executed"})
        return
    case "download_image":
        // Expect { url: string, destName: string }
        var payload struct {
            URL      string `json:"url"`
            DestName string `json:"destName"`
        }
        if err := json.Unmarshal(req.Data, &payload); err != nil {
            http.Error(w, fmt.Sprintf("invalid data: %v", err), http.StatusBadRequest)
            return
        }
        if payload.URL == "" || payload.DestName == "" {
            http.Error(w, "missing url or destName", http.StatusBadRequest)
            return
        }
        out, err := downloadAndDecompress(payload.URL, payload.DestName)
        if err != nil {
            http.Error(w, fmt.Sprintf("download failed: %v", err), http.StatusBadGateway)
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{"result": "executed", "path": out})
        return

    case "run_vm":
        // Expect { id: string } or { name: string }
        var payload map[string]string
        _ = json.Unmarshal(req.Data, &payload)
        name := payload["id"]
        if name == "" {
            name = payload["name"]
        }
        if name == "" {
            http.Error(w, "missing id/name", http.StatusBadRequest)
            return
        }
        if err := execTart("run", name); err != nil {
            http.Error(w, fmt.Sprintf("tart run failed: %v", err), http.StatusBadGateway)
            return
        }
        json.NewEncoder(w).Encode(map[string]string{"result": "executed"})
        return

    case "delete_vm":
        var payload map[string]string
        _ = json.Unmarshal(req.Data, &payload)
        name := payload["id"]
        if name == "" {
            name = payload["name"]
        }
        if name == "" {
            http.Error(w, "missing id/name", http.StatusBadRequest)
            return
        }
        if err := execTart("delete", name); err != nil {
            http.Error(w, fmt.Sprintf("tart delete failed: %v", err), http.StatusBadGateway)
            return
        }
        json.NewEncoder(w).Encode(map[string]string{"result": "executed"})
        return

    case "create_vm":
        // Validate that a decompressed image exists in cache for the VM name; if only .img.xz exists and xz is available, decompress it now.
        // Expect { name: string, image: string }
        var payload struct {
            Name  string `json:"name"`
            Image string `json:"image"`
        }
        if err := json.Unmarshal(req.Data, &payload); err != nil {
            http.Error(w, fmt.Sprintf("invalid data: %v", err), http.StatusBadRequest)
            return
        }
        if payload.Name == "" {
            http.Error(w, "missing name", http.StatusBadRequest)
            return
        }
        cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "tart-images")
        imgPath := filepath.Join(cacheDir, payload.Name+".img")
        if _, err := os.Stat(imgPath); err != nil {
            // attempt on-demand decompress if .img.xz exists and xz is available
            xzPath := filepath.Join(cacheDir, payload.Name+".img.xz")
            if _, err2 := os.Stat(xzPath); err2 == nil {
                if _, err3 := exec.LookPath("xz"); err3 == nil {
                    cmd := exec.Command("xz", "-d", "-k", "-f", xzPath)
                    cmd.Stdout = os.Stdout
                    cmd.Stderr = os.Stderr
                    if err4 := cmd.Run(); err4 != nil {
                        http.Error(w, fmt.Sprintf("auto-decompress failed: %v", err4), http.StatusBadGateway)
                        return
                    }
                    // re-check
                    if _, err5 := os.Stat(imgPath); err5 != nil {
                        http.Error(w, fmt.Sprintf("decompressed image missing: %s", imgPath), http.StatusBadGateway)
                        return
                    }
                } else {
                    http.Error(w, "xz not installed to decompress .img.xz", http.StatusBadGateway)
                    return
                }
            } else {
                http.Error(w, fmt.Sprintf("image not found: %s (nor %s)", imgPath, xzPath), http.StatusBadGateway)
                return
            }
        }
        // In a future iteration, this is where we'd call Tart to import/create a VM from the disk image.
        json.NewEncoder(w).Encode(map[string]string{"result": "executed", "image": imgPath})
        return

    default:
        http.Error(w, "unknown action", http.StatusBadRequest)
        return
    }
}

func execTart(subcmd string, args ...string) error {
    if _, err := exec.LookPath("tart"); err != nil {
        return errors.New("tart binary not found in PATH")
    }
    cmd := exec.Command("tart", append([]string{subcmd}, args...)...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func downloadAndDecompress(url, destName string) (string, error) {
    cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "tart-images")
    if err := os.MkdirAll(cacheDir, 0o755); err != nil {
        return "", err
    }
    xzPath := filepath.Join(cacheDir, destName+".img.xz")
    imgPath := filepath.Join(cacheDir, destName+".img")

    // Download
    if err := httpDownload(url, xzPath); err != nil {
        return "", err
    }
    // Decompress using external xz, if available
    if _, err := exec.LookPath("xz"); err == nil {
        // xz -d -k -f <file.xz>
        cmd := exec.Command("xz", "-d", "-k", "-f", xzPath)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            return "", fmt.Errorf("xz decompress failed: %w", err)
        }
    } else {
        // No xz available; return the .xz path and let caller handle
        return xzPath, nil
    }
    // Wait a moment for file system to settle
    time.Sleep(100 * time.Millisecond)
    if _, err := os.Stat(imgPath); err == nil {
        return imgPath, nil
    }
    // Fallback to xz path if decompressed not found
    return xzPath, nil
}

func httpDownload(url, dest string) error {
    out, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer out.Close()
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("download http status: %s", resp.Status)
    }
    if _, err := io.Copy(out, resp.Body); err != nil {
        return err
    }
    // Basic sanity: non-empty file
    fi, _ := out.Stat()
    if fi == nil || fi.Size() == 0 {
        return errors.New("downloaded file is empty")
    }
    return nil
}
