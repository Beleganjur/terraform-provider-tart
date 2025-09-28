//go:build integration
// +build integration

package tests

import (
    "bytes"
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"
    "testing"
)

func mustLookPath(t *testing.T, bin string) string {
    t.Helper()
    p, err := exec.LookPath(bin)
    if err != nil {
        t.Skipf("%s not found in PATH; skipping integration test", bin)
        return ""
    }
    return p
}

func runCmd(t *testing.T, dir string, name string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// TestTerraformAffectsTart asserts that a terraform apply leads to observable changes
// in `tart list` output (best-effort). This test is intended to catch no-op controller/executor flows.
func TestTerraformAffectsTart(t *testing.T) {
	// Preconditions
	_ = mustLookPath(t, "terraform")
	_ = mustLookPath(t, "tart")
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("tart is only available on Apple Silicon hosts; skipping")
	}

	// Project root (where main.tf lives)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	projDir := filepath.Clean(filepath.Join(wd, ".."))

	// Ensure clean state (best-effort)
	_, _, _ = runCmd(t, projDir, "make", "clean")

	// Init and plan/apply
	if _, stderr, err := runCmd(t, projDir, "terraform", "init", "-upgrade", "-input=false"); err != nil {
		t.Fatalf("terraform init failed: %v\n%s", err, stderr)
	}
	if _, stderr, err := runCmd(t, projDir, "terraform", "apply", "-auto-approve"); err != nil {
		t.Fatalf("terraform apply failed: %v\n%s", err, stderr)
	}

	// Read outputs to learn desired VM names
	stdout, stderr, err := runCmd(t, projDir, "terraform", "output", "-json")
	if err != nil {
		t.Fatalf("terraform output failed: %v\n%s", err, stderr)
	}

	// Parse JSON outputs robustly
	var tfOut map[string]struct {
		Sensitive bool        `json:"sensitive"`
		Type      interface{} `json:"type"`
		Value     interface{} `json:"value"`
	}
	if err := json.Unmarshal([]byte(stdout), &tfOut); err != nil {
		t.Fatalf("failed parsing terraform outputs JSON: %v\nraw: %s", err, stdout)
	}

	// Extract expected VM names (string values)
	expectedNames := make([]string, 0, 2)
	for _, key := range []string{"debian_13_vm_name", "sequoia_base_vm_name"} {
		if v, ok := tfOut[key]; ok {
			if s, ok2 := v.Value.(string); ok2 && s != "" {
				expectedNames = append(expectedNames, s)
			}
		}
	}
	if len(expectedNames) == 0 {
		t.Fatalf("no VM names found in terraform outputs; outputs: %s", stdout)
	}

	// Check `tart list` includes at least one of the expected VM names
	listOut, listErr, err := runCmd(t, projDir, "tart", "list")
	if err != nil {
		t.Fatalf("tart list failed: %v\n%s", err, listErr)
	}

	matched := false
	for _, name := range expectedNames {
		if name != "" && strings.Contains(listOut, name) {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("No Terraform-managed VM names (%s) were found in 'tart list' output.\nterraform outputs (json):\n%s\n---\n`tart list` output:\n%s",
			strings.Join(expectedNames, ", "), stdout, listOut)
	}
}
