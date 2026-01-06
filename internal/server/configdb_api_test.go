package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/googleapis/genai-toolbox/internal/log"
	"github.com/googleapis/genai-toolbox/internal/server/resources"
	"github.com/googleapis/genai-toolbox/internal/telemetry"
)

func TestConfigDBSourceCRUD(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "toolbox-configdb-*")
	if err != nil {
		t.Fatalf("unable to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	dbPath := filepath.Join(tmpDir, "config.db")
	dbPathEsc := url.QueryEscape(dbPath)

	testLogger, err := log.NewStdLogger(os.Stdout, os.Stderr, "info")
	if err != nil {
		t.Fatalf("unable to initialize logger: %s", err)
	}

	otelShutdown, err := telemetry.SetupOTel(ctx, fakeVersionString, "", false, "toolbox")
	if err != nil {
		t.Fatalf("unable to setup otel: %s", err)
	}
	defer func() {
		_ = otelShutdown(ctx)
	}()

	instrumentation, err := telemetry.CreateTelemetryInstrumentation(fakeVersionString)
	if err != nil {
		t.Fatalf("unable to create custom metrics: %s", err)
	}

	server := Server{
		version:         fakeVersionString,
		logger:          testLogger,
		instrumentation: instrumentation,
		sseManager:      newSseManager(ctx),
		ResourceMgr:     resources.NewResourceManager(nil, nil, nil, nil, nil, nil),
	}

	r, err := apiRouter(&server)
	if err != nil {
		t.Fatalf("unable to initialize api router: %s", err)
	}
	ts := runServer(r, false)
	defer ts.Close()

	// 初始 list：db 不存在时返回空列表
	{
		resp, body, err := runRequest(ts, http.MethodGet, "/config/sources?dbPath="+dbPathEsc, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", resp.StatusCode, string(body))
		}
		var got []configDBSourceDTO
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unable to decode response: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty list, got %d", len(got))
		}
	}

	// Create
	{
		reqBody, _ := json.Marshal(configDBCreateSourceRequest{
			Name: "db1",
			Kind: "postgres",
			Config: map[string]any{
				"kind":     "postgres",
				"host":     "localhost",
				"port":     5432,
				"database": "testdb",
			},
		})
		resp, body, err := runRequest(ts, http.MethodPost, "/config/sources?dbPath="+dbPathEsc, bytes.NewReader(reqBody), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("unexpected status: %d body=%s", resp.StatusCode, string(body))
		}
	}

	// Conflict create same
	{
		reqBody, _ := json.Marshal(configDBCreateSourceRequest{Name: "db1", Kind: "postgres", Config: map[string]any{}})
		resp, _, err := runRequest(ts, http.MethodPost, "/config/sources?dbPath="+dbPathEsc, bytes.NewReader(reqBody), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	}

	// Get
	{
		resp, body, err := runRequest(ts, http.MethodGet, "/config/sources/db1?dbPath="+dbPathEsc, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", resp.StatusCode, string(body))
		}
		var got configDBSourceDTO
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unable to decode response: %v", err)
		}
		if got.Name != "db1" || got.Kind != "postgres" {
			t.Fatalf("unexpected dto: %+v", got)
		}
	}

	// Update
	{
		reqBody, _ := json.Marshal(configDBUpdateSourceRequest{
			Kind: "postgres",
			Config: map[string]any{
				"kind": "postgres",
				"host": "127.0.0.1",
				"port": 5432,
			},
		})
		resp, body, err := runRequest(ts, http.MethodPut, "/config/sources/db1?dbPath="+dbPathEsc, bytes.NewReader(reqBody), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", resp.StatusCode, string(body))
		}
	}

	// Delete
	{
		resp, _, err := runRequest(ts, http.MethodDelete, "/config/sources/db1?dbPath="+dbPathEsc, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}

	// Get after delete -> 404
	{
		resp, _, err := runRequest(ts, http.MethodGet, "/config/sources/db1?dbPath="+dbPathEsc, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	}
}

