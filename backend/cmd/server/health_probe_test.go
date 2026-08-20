package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsHealthProbe(t *testing.T) {
	assert.True(t, isHealthProbe([]string{"/server", "-healthcheck"}))
	assert.False(t, isHealthProbe([]string{"/server"}))
	assert.False(t, isHealthProbe([]string{"/server", "-other"}))
}

func TestHealthProbeURL(t *testing.T) {
	t.Setenv("PORT", "")
	assert.Equal(t, "http://127.0.0.1:8080/api/v2/health", healthProbeURL())

	t.Setenv("PORT", "9999")
	assert.Equal(t, "http://127.0.0.1:9999/api/v2/health", healthProbeURL())
}

func TestProbeHealth(t *testing.T) {
	t.Run("2xx なら nil", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		require.NoError(t, probeHealth(context.Background(), srv.URL))
	})

	t.Run("DB 断の 503 はエラー", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer srv.Close()

		assert.ErrorContains(t, probeHealth(context.Background(), srv.URL), "503")
	})

	t.Run("接続できないときはエラー", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := srv.URL
		srv.Close()

		assert.Error(t, probeHealth(context.Background(), url))
	})
}
