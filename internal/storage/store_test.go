// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	// Import source types for registration
	_ "github.com/googleapis/genai-toolbox/internal/sources/postgres"
)

func TestOpenAndClose(t *testing.T) {
	ctx := context.Background()

	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "toolbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open the store
	store, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}

	// Check if the database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Check that the database is empty initially
	isEmpty, err := store.IsEmpty(ctx)
	if err != nil {
		t.Fatalf("Failed to check if empty: %v", err)
	}
	if !isEmpty {
		t.Error("Expected database to be empty initially")
	}

	// Close the store
	if err := store.Close(); err != nil {
		t.Fatalf("Failed to close store: %v", err)
	}
}

func TestSaveAndLoadSource(t *testing.T) {
	ctx := context.Background()

	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "toolbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open the store
	store, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	// Save a source
	sourceConfig := map[string]any{
		"kind":     "postgres",
		"host":     "localhost",
		"port":     "5432",
		"database": "testdb",
		"user":     "testuser",
		"password": "testpass",
	}
	err = store.SaveSource(ctx, "test-source", "postgres", sourceConfig)
	if err != nil {
		t.Fatalf("Failed to save source: %v", err)
	}

	// Check that the database is no longer empty
	isEmpty, err := store.IsEmpty(ctx)
	if err != nil {
		t.Fatalf("Failed to check if empty: %v", err)
	}
	if isEmpty {
		t.Error("Expected database to not be empty after saving source")
	}

	// Load the data
	data, err := store.LoadToolsFileData(ctx)
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	if len(data.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(data.Sources))
	}

	if _, ok := data.Sources["test-source"]; !ok {
		t.Error("Expected to find 'test-source' in loaded sources")
	}
}

func TestExists(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "toolbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Check that the database doesn't exist initially
	exists, err := Exists(dbPath)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected database to not exist initially")
	}

	// Create the database
	ctx := context.Background()
	store, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	store.Close()

	// Check that the database now exists
	exists, err = Exists(dbPath)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected database to exist after creation")
	}
}

func TestMergeConfigs(t *testing.T) {
	// Test merge with nil base
	override := &ConfigData{}
	result := MergeConfigs(nil, override)
	if result != override {
		t.Error("Expected override to be returned when base is nil")
	}

	// Test merge with nil override
	base := &ConfigData{}
	result = MergeConfigs(base, nil)
	if result != base {
		t.Error("Expected base to be returned when override is nil")
	}

	// Test merge with both nil
	result = MergeConfigs(nil, nil)
	if result != nil {
		t.Error("Expected nil when both base and override are nil")
	}
}

