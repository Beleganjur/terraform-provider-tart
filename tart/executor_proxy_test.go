package tart

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestForwardToExecutor_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"executed"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	old := os.Getenv("EXECUTOR_URL")
	os.Setenv("EXECUTOR_URL", srv.URL)
	defer func() {
		if old == "" { os.Unsetenv("EXECUTOR_URL") } else { os.Setenv("EXECUTOR_URL", old) }
	}()

	if err := forwardToExecutor("clone_vm", map[string]string{"name":"n","image":"i"}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestForwardToExecutor_Non200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"bad"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	old := os.Getenv("EXECUTOR_URL")
	os.Setenv("EXECUTOR_URL", srv.URL)
	defer func() {
		if old == "" { os.Unsetenv("EXECUTOR_URL") } else { os.Setenv("EXECUTOR_URL", old) }
	}()

	if err := forwardToExecutor("pull_image", map[string]string{"ref":"x"}); err == nil {
		t.Fatalf("expected error on non-200 status")
	}
}
