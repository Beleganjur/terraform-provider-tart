package tart

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
)

// cliCreateVM creates a new VM by cloning a base image using the Tart CLI.
func cliCreateVM(ctx context.Context, baseImage, newVMName string) error {
    cmd := exec.CommandContext(ctx, "tart", "clone", baseImage, newVMName)
    out, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("tart clone failed: %v\nOutput: %s", err, string(out))
    }
    return nil
}

// cliGetVM checks a VM's existence by listing via the Tart CLI.
func cliGetVM(ctx context.Context, vmName string) (bool, error) {
    cmd := exec.CommandContext(ctx, "tart", "list")
    out, err := cmd.Output()
    if err != nil {
        return false, fmt.Errorf("tart list failed: %w", err)
    }
    // Output format: may need more robust parsing
    list := string(out)
    return strings.Contains(list, vmName), nil
}

// cliDeleteVM deletes a VM using the Tart CLI.
func cliDeleteVM(ctx context.Context, vmName string) error {
    cmd := exec.CommandContext(ctx, "tart", "delete", vmName)
    out, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("tart delete failed: %v\nOutput: %s", err, out)
    }
    return nil
}

