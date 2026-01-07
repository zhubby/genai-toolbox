package server

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/googleapis/genai-toolbox/internal/auth"
	"github.com/googleapis/genai-toolbox/internal/log"
	"github.com/googleapis/genai-toolbox/internal/prompts"
	"github.com/googleapis/genai-toolbox/internal/server/resources"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"github.com/googleapis/genai-toolbox/internal/sources/sqlite"
	"github.com/googleapis/genai-toolbox/internal/storage"
	"github.com/googleapis/genai-toolbox/internal/telemetry"
	"github.com/googleapis/genai-toolbox/internal/tools"
)

func setUpAPIRouterWithSources(t *testing.T, sourcesMap map[string]sources.Source) (chi.Router, func()) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	testLogger, err := log.NewStdLogger(os.Stdout, os.Stderr, "info")
	if err != nil {
		t.Fatalf("unable to initialize logger: %s", err)
	}

	otelShutdown, err := telemetry.SetupOTel(ctx, fakeVersionString, "", false, "toolbox")
	if err != nil {
		t.Fatalf("unable to setup otel: %s", err)
	}
	instrumentation, err := telemetry.CreateTelemetryInstrumentation(fakeVersionString)
	if err != nil {
		t.Fatalf("unable to create custom metrics: %s", err)
	}

	sseManager := newSseManager(ctx)

	rm := resources.NewResourceManager(sourcesMap, map[string]auth.AuthService{}, map[string]tools.Tool{}, map[string]tools.Toolset{}, map[string]prompts.Prompt{}, map[string]prompts.Promptset{})
	server := &Server{
		version:         fakeVersionString,
		logger:          testLogger,
		instrumentation: instrumentation,
		sseManager:      sseManager,
		ResourceMgr:     rm,
		configDBPath:    filepath.Join(t.TempDir(), "config.db"),
	}

	r, err := apiRouter(server)
	if err != nil {
		t.Fatalf("unable to initialize api router: %s", err)
	}

	shutdown := func() {
		cancel()
		if err := otelShutdown(ctx); err != nil {
			t.Fatalf("error shutting down OpenTelemetry: %s", err)
		}
	}

	return r, shutdown
}

func TestSQLExecute_SQLite_SelectOne(t *testing.T) {
	ctx := context.Background()

	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	cfg := sqlite.Config{Name: "db1", Kind: "sqlite", Database: dbPath}
	inst, err := telemetry.CreateTelemetryInstrumentation(fakeVersionString)
	if err != nil {
		t.Fatalf("create telemetry instrumentation: %v", err)
	}
	src, err := cfg.Initialize(ctx, inst.Tracer)
	if err != nil {
		t.Fatalf("init sqlite source: %v", err)
	}

	r, shutdown := setUpAPIRouterWithSources(t, map[string]sources.Source{"db1": src})
	defer shutdown()

	ts := runServer(r, false)
	defer ts.Close()

	body := bytes.NewBufferString(`{"source":"db1","statement":"SELECT 1 AS a","parameters":[]}`)
	resp, respBody, err := runRequest(ts, "POST", "/sql/execute", body, nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var out struct {
		Columns  []string `json:"columns"`
		Rows     [][]any  `json:"rows"`
		RowCount int      `json:"rowCount"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, string(respBody))
	}
	if len(out.Columns) != 1 || out.Columns[0] != "a" {
		t.Fatalf("unexpected columns: %#v", out.Columns)
	}
	if out.RowCount != 1 || len(out.Rows) != 1 || len(out.Rows[0]) != 1 {
		t.Fatalf("unexpected rows: rowCount=%d rows=%#v", out.RowCount, out.Rows)
	}
	if v, ok := out.Rows[0][0].(float64); !ok || v != 1 {
		t.Fatalf("unexpected cell value: %#v", out.Rows[0][0])
	}
}

func TestSQLExecute_SQLite_EmptyRowsStillReturnsColumns(t *testing.T) {
	ctx := context.Background()

	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	cfg := sqlite.Config{Name: "db1", Kind: "sqlite", Database: dbPath}
	inst, err := telemetry.CreateTelemetryInstrumentation(fakeVersionString)
	if err != nil {
		t.Fatalf("create telemetry instrumentation: %v", err)
	}
	src, err := cfg.Initialize(ctx, inst.Tracer)
	if err != nil {
		t.Fatalf("init sqlite source: %v", err)
	}

	r, shutdown := setUpAPIRouterWithSources(t, map[string]sources.Source{"db1": src})
	defer shutdown()

	ts := runServer(r, false)
	defer ts.Close()

	body := bytes.NewBufferString(`{"source":"db1","statement":"SELECT 1 AS a WHERE 0","parameters":[]}`)
	resp, respBody, err := runRequest(ts, "POST", "/sql/execute", body, nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var out struct {
		Columns  []string `json:"columns"`
		Rows     [][]any  `json:"rows"`
		RowCount int      `json:"rowCount"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, string(respBody))
	}
	// Current backend response is "empty table" when there are no rows.
	// (Most sources return only row objects; when there are no rows, column names
	// are not available without additional metadata.)
	if out.RowCount != 0 || len(out.Rows) != 0 || len(out.Columns) != 0 {
		t.Fatalf("unexpected rows: rowCount=%d rows=%#v", out.RowCount, out.Rows)
	}
}

func TestSQLExecute_LoadSourceFromConfigDB_WhenNotInResourceMgr(t *testing.T) {
	ctx := context.Background()

	// Create a config DB and persist a sqlite source named "local".
	dbPath := filepath.Join(t.TempDir(), "config.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("open config db: %v", err)
	}
	_, err = store.CreateSource(ctx, "local", "sqlite", map[string]any{
		"kind":     "sqlite",
		"name":     "local",
		"database": filepath.Join(t.TempDir(), "local.sqlite"),
	})
	if err != nil {
		_ = store.Close()
		t.Fatalf("create source: %v", err)
	}
	_ = store.Close()

	// Server has empty ResourceMgr sources, but configDBPath points to dbPath.
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()
	inst, err := telemetry.CreateTelemetryInstrumentation(fakeVersionString)
	if err != nil {
		t.Fatalf("create telemetry instrumentation: %v", err)
	}
	testLogger, err := log.NewStdLogger(os.Stdout, os.Stderr, "info")
	if err != nil {
		t.Fatalf("unable to initialize logger: %s", err)
	}
	rm := resources.NewResourceManagerForDB(dbPath, inst.Tracer, nil, fakeVersionString)
	srv := &Server{
		version:         fakeVersionString,
		logger:          testLogger,
		instrumentation: inst,
		sseManager:      newSseManager(ctx2),
		ResourceMgr:     rm,
		configDBPath:    dbPath,
	}
	r, err := apiRouter(srv)
	if err != nil {
		t.Fatalf("api router: %v", err)
	}
	ts := runServer(r, false)
	defer ts.Close()

	body := bytes.NewBufferString(`{"source":"local","statement":"SELECT 1 AS a","parameters":[]}`)
	resp, respBody, err := runRequest(ts, "POST", "/sql/execute", body, nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d body=%s", resp.StatusCode, string(respBody))
	}
}


