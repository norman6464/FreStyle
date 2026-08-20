package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	// healthcheck の成否は 2xx かどうかだけで決まる。境界(200-299)を固定しておく。
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{name: "200 は成功", status: http.StatusOK},
		{name: "204 は成功", status: http.StatusNoContent},
		{name: "299 は成功", status: 299},
		{name: "300 はエラー", status: http.StatusMultipleChoices, wantErr: true},
		{name: "DB 断の 503 はエラー", status: http.StatusServiceUnavailable, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			err := probeHealth(context.Background(), srv.URL)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, strconv.Itoa(tt.status))
		})
	}

	t.Run("接続できないときはエラー", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := srv.URL
		srv.Close()

		assert.Error(t, probeHealth(context.Background(), url))
	})
}
