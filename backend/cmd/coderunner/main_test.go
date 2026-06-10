package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/sandbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer() *httptest.Server {
	return httptest.NewServer(newMux(sandbox.NewRunner()))
}

func TestServer_Healthz(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServer_Run_PHP(t *testing.T) {
	if _, err := exec.LookPath("php"); err != nil {
		t.Skip("php not found in PATH")
	}
	srv := newTestServer()
	defer srv.Close()

	body, _ := json.Marshal(domain.CodeExecutionInput{Code: `<?php echo "hi";`, Language: "php"})
	resp, err := http.Post(srv.URL+"/run", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out domain.CodeExecutionResult
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	assert.Equal(t, "hi", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

func TestServer_Run_InvalidBody(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/run", "application/json", bytes.NewReader([]byte("not-json")))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestServer_Warmup(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	body, _ := json.Marshal(map[string]string{"language": "php"})
	resp, err := http.Post(srv.URL+"/warmup", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	assert.True(t, out["ready"])
}
