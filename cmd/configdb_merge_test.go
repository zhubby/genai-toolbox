package cmd

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/googleapis/genai-toolbox/internal/sources"
	httpsrc "github.com/googleapis/genai-toolbox/internal/sources/http"
	"github.com/googleapis/genai-toolbox/internal/storage"
)

func TestResolveConfigDBPath_DefaultsForYAML(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := resolveConfigDBPath(true, true, "")
	if err != nil {
		t.Fatalf("resolveConfigDBPath() err = %v", err)
	}
	want := filepath.Join(home, ".toolbox", "config.db")
	if got != want {
		t.Fatalf("resolveConfigDBPath() = %q, want %q", got, want)
	}
}

func TestMergeAndPersistToDB_YAMLOverridesDB(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "config.db")

	base := &storage.ConfigData{
		SourceConfigs: map[string]sources.SourceConfig{
			"s1": httpsrc.Config{
				Name:    "s1",
				Kind:    httpsrc.SourceKind,
				BaseURL: "http://base.example",
				Timeout: "30s",
			},
		},
	}
	if err := saveMergedConfigToDB(ctx, dbPath, base); err != nil {
		t.Fatalf("saveMergedConfigToDB(base) err = %v", err)
	}

	override := &storage.ConfigData{
		SourceConfigs: map[string]sources.SourceConfig{
			"s1": httpsrc.Config{
				Name:    "s1",
				Kind:    httpsrc.SourceKind,
				BaseURL: "http://override.example",
				Timeout: "30s",
			},
		},
	}
	if err := mergeAndPersistToDB(ctx, dbPath, override); err != nil {
		t.Fatalf("mergeAndPersistToDB(override) err = %v", err)
	}

	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("storage.Open err = %v", err)
	}
	defer store.Close() //nolint:errcheck

	data, err := store.LoadToolsFileData(ctx)
	if err != nil {
		t.Fatalf("store.LoadToolsFileData err = %v", err)
	}

	cfgAny, ok := data.Sources["s1"]
	if !ok {
		t.Fatalf("expected source %q in db", "s1")
	}
	cfg, ok := cfgAny.(httpsrc.Config)
	if !ok {
		t.Fatalf("expected source config type %T, got %T", httpsrc.Config{}, cfgAny)
	}
	if cfg.BaseURL != "http://override.example" {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, "http://override.example")
	}
}


