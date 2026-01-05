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
	"fmt"

	"github.com/googleapis/genai-toolbox/internal/auth"
	"github.com/googleapis/genai-toolbox/internal/prompts"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"github.com/googleapis/genai-toolbox/internal/tools"
)

// ConfigData holds all configuration data in a format compatible with server.ServerConfig.
// This is used to transfer data from storage to the command layer without import cycles.
type ConfigData struct {
	SourceConfigs      map[string]sources.SourceConfig
	AuthServiceConfigs map[string]auth.AuthServiceConfig
	ToolConfigs        map[string]tools.ToolConfig
	ToolsetConfigs     map[string]tools.ToolsetConfig
	PromptConfigs      map[string]prompts.PromptConfig
}

// LoadFromDB loads all configurations from the SQLite database.
// If dbPath is empty, it uses the default path.
// Returns nil if the database doesn't exist or is empty.
func LoadFromDB(ctx context.Context, dbPath string) (*ConfigData, error) {
	// Check if database exists
	exists, err := Exists(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}
	if !exists {
		return nil, nil
	}

	// Open database
	store, err := Open(ctx, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	// Check if database is empty
	isEmpty, err := store.IsEmpty(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database is empty: %w", err)
	}
	if isEmpty {
		return nil, nil
	}

	// Load data
	data, err := store.LoadToolsFileData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load data from database: %w", err)
	}

	return &ConfigData{
		SourceConfigs:      data.Sources,
		AuthServiceConfigs: data.AuthServices,
		ToolConfigs:        data.Tools,
		ToolsetConfigs:     data.Toolsets,
		PromptConfigs:      data.Prompts,
	}, nil
}

// MergeConfigs merges two ConfigData instances.
// The 'override' config takes precedence over 'base' for conflicting keys.
func MergeConfigs(base, override *ConfigData) *ConfigData {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := &ConfigData{
		SourceConfigs:      make(map[string]sources.SourceConfig),
		AuthServiceConfigs: make(map[string]auth.AuthServiceConfig),
		ToolConfigs:        make(map[string]tools.ToolConfig),
		ToolsetConfigs:     make(map[string]tools.ToolsetConfig),
		PromptConfigs:      make(map[string]prompts.PromptConfig),
	}

	// Copy base configs
	for k, v := range base.SourceConfigs {
		result.SourceConfigs[k] = v
	}
	for k, v := range base.AuthServiceConfigs {
		result.AuthServiceConfigs[k] = v
	}
	for k, v := range base.ToolConfigs {
		result.ToolConfigs[k] = v
	}
	for k, v := range base.ToolsetConfigs {
		result.ToolsetConfigs[k] = v
	}
	for k, v := range base.PromptConfigs {
		result.PromptConfigs[k] = v
	}

	// Override with override configs
	for k, v := range override.SourceConfigs {
		result.SourceConfigs[k] = v
	}
	for k, v := range override.AuthServiceConfigs {
		result.AuthServiceConfigs[k] = v
	}
	for k, v := range override.ToolConfigs {
		result.ToolConfigs[k] = v
	}
	for k, v := range override.ToolsetConfigs {
		result.ToolsetConfigs[k] = v
	}
	for k, v := range override.PromptConfigs {
		result.PromptConfigs[k] = v
	}

	return result
}

// HasAnyConfig returns true if the ConfigData contains any configuration.
func (c *ConfigData) HasAnyConfig() bool {
	if c == nil {
		return false
	}
	return len(c.SourceConfigs) > 0 ||
		len(c.AuthServiceConfigs) > 0 ||
		len(c.ToolConfigs) > 0 ||
		len(c.ToolsetConfigs) > 0 ||
		len(c.PromptConfigs) > 0
}

