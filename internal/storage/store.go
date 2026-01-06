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

// Package storage provides configuration storage functionality using SQLite.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/googleapis/genai-toolbox/internal/storage/ent"

	// Import SQLite driver (pure Go implementation)
	_ "modernc.org/sqlite"
)

// ent schema migration (client.Schema.Create) is not safe to run concurrently.
// Since Toolbox may open the config DB from multiple HTTP requests at once,
// we serialize migrations to prevent "concurrent map writes" panics.
var schemaCreateMu sync.Mutex

// Store provides an interface for configuration storage operations.
type Store struct {
	client *ent.Client
	dbPath string
}

// DefaultDBPath returns the default database path (~/.toolbox/config.db).
func DefaultDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".toolbox", "config.db"), nil
}

// Open opens or creates a SQLite database at the specified path.
// If the path is empty, it uses the default path (~/.toolbox/config.db).
func Open(ctx context.Context, dbPath string) (*Store, error) {
	var err error
	if dbPath == "" {
		dbPath, err = DefaultDBPath()
		if err != nil {
			return nil, err
		}
	}

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open the database using ent with SQLite
	// Use modernc.org/sqlite which is a pure Go driver registered as "sqlite"
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys (required by ent)
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create ent driver from sql.DB
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	// Run auto migration to create/update schema
	// Use WithForeignKeys(false) to skip the foreign key check during migration
	schemaCreateMu.Lock()
	err = client.Schema.Create(ctx)
	schemaCreateMu.Unlock()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &Store{
		client: client,
		dbPath: dbPath,
	}, nil
}

// Exists checks if a database file exists at the specified path.
// If the path is empty, it checks the default path.
func Exists(dbPath string) (bool, error) {
	var err error
	if dbPath == "" {
		dbPath, err = DefaultDBPath()
		if err != nil {
			return false, err
		}
	}

	_, err = os.Stat(dbPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Client returns the underlying ent client.
func (s *Store) Client() *ent.Client {
	return s.client
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// DBPath returns the path to the database file.
func (s *Store) DBPath() string {
	return s.dbPath
}

// IsEmpty checks if the database has any configurations stored.
func (s *Store) IsEmpty(ctx context.Context) (bool, error) {
	sourceCount, err := s.client.Source.Query().Count(ctx)
	if err != nil {
		return false, err
	}
	toolCount, err := s.client.Tool.Query().Count(ctx)
	if err != nil {
		return false, err
	}
	promptCount, err := s.client.Prompt.Query().Count(ctx)
	if err != nil {
		return false, err
	}

	return sourceCount == 0 && toolCount == 0 && promptCount == 0, nil
}
