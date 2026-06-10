package coderunner_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/coderunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Run_SendsInputAndDecodesResult(t *testing.T) {
	var gotPath string
	var gotBody domain.CodeExecutionInput
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(domain.CodeExecutionResult{Stdout: "hi\n", ExitCode: 0})
	}))
	defer srv.Close()

	c := coderunner.NewClient(srv.URL)
	out, err := c.Run(context.Background(), domain.CodeExecutionInput{
		Code: `<?php echo "hi";`, Language: "php", Stdin: "x",
	})
	require.NoError(t, err)
	assert.Equal(t, "/run", gotPath)
	assert.Equal(t, "php", gotBody.Language)
	assert.Equal(t, "x", gotBody.Stdin)
	assert.Equal(t, "hi\n", out.Stdout)
}

func TestClient_Warmup_SendsLanguage(t *testing.T) {
	var gotPath string
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(map[string]bool{"ready": true})
	}))
	defer srv.Close()

	c := coderunner.NewClient(srv.URL)
	require.NoError(t, c.Warmup(context.Background(), "go"))
	assert.Equal(t, "/warmup", gotPath)
	assert.Equal(t, "go", gotBody["language"])
}

func TestClient_Run_ErrorsOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := coderunner.NewClient(srv.URL)
	_, err := c.Run(context.Background(), domain.CodeExecutionInput{Language: "go"})
	assert.Error(t, err)
}
