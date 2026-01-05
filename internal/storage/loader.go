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
	"github.com/googleapis/genai-toolbox/internal/auth/google"
	"github.com/googleapis/genai-toolbox/internal/prompts"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"github.com/googleapis/genai-toolbox/internal/storage/ent"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"github.com/googleapis/genai-toolbox/internal/util"
)

// ToolsFileData represents the configuration data that can be loaded from storage.
// This mirrors the structure used in cmd.ToolsFile but without import cycles.
type ToolsFileData struct {
	Sources      map[string]sources.SourceConfig
	AuthServices map[string]auth.AuthServiceConfig
	Tools        map[string]tools.ToolConfig
	Toolsets     map[string]tools.ToolsetConfig
	Prompts      map[string]prompts.PromptConfig
}

// LoadToolsFileData loads all configurations from the database and returns them
// as decoded config objects ready to be used.
func (s *Store) LoadToolsFileData(ctx context.Context) (*ToolsFileData, error) {
	data := &ToolsFileData{
		Sources:      make(map[string]sources.SourceConfig),
		AuthServices: make(map[string]auth.AuthServiceConfig),
		Tools:        make(map[string]tools.ToolConfig),
		Toolsets:     make(map[string]tools.ToolsetConfig),
		Prompts:      make(map[string]prompts.PromptConfig),
	}

	// Load sources
	if err := s.loadSources(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	// Load auth services
	if err := s.loadAuthServices(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to load auth services: %w", err)
	}

	// Load tools
	if err := s.loadTools(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to load tools: %w", err)
	}

	// Load toolsets
	if err := s.loadToolsets(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to load toolsets: %w", err)
	}

	// Load prompts
	if err := s.loadPrompts(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to load prompts: %w", err)
	}

	return data, nil
}

func (s *Store) loadSources(ctx context.Context, data *ToolsFileData) error {
	sources_, err := s.client.Source.Query().All(ctx)
	if err != nil {
		return err
	}

	for _, src := range sources_ {
		config, err := decodeSourceConfig(ctx, src)
		if err != nil {
			return fmt.Errorf("failed to decode source %q: %w", src.ID, err)
		}
		data.Sources[src.ID] = config
	}
	return nil
}

func (s *Store) loadAuthServices(ctx context.Context, data *ToolsFileData) error {
	authServices, err := s.client.AuthService.Query().All(ctx)
	if err != nil {
		return err
	}

	for _, as := range authServices {
		config, err := decodeAuthServiceConfig(ctx, as)
		if err != nil {
			return fmt.Errorf("failed to decode auth service %q: %w", as.ID, err)
		}
		data.AuthServices[as.ID] = config
	}
	return nil
}

func (s *Store) loadTools(ctx context.Context, data *ToolsFileData) error {
	tools_, err := s.client.Tool.Query().All(ctx)
	if err != nil {
		return err
	}

	for _, t := range tools_ {
		config, err := decodeToolConfig(ctx, t)
		if err != nil {
			return fmt.Errorf("failed to decode tool %q: %w", t.ID, err)
		}
		data.Tools[t.ID] = config
	}
	return nil
}

func (s *Store) loadToolsets(ctx context.Context, data *ToolsFileData) error {
	toolsets, err := s.client.Toolset.Query().All(ctx)
	if err != nil {
		return err
	}

	for _, ts := range toolsets {
		data.Toolsets[ts.ID] = tools.ToolsetConfig{
			Name:      ts.ID,
			ToolNames: ts.ToolNames,
		}
	}
	return nil
}

func (s *Store) loadPrompts(ctx context.Context, data *ToolsFileData) error {
	prompts_, err := s.client.Prompt.Query().All(ctx)
	if err != nil {
		return err
	}

	for _, p := range prompts_ {
		config, err := decodePromptConfig(ctx, p)
		if err != nil {
			return fmt.Errorf("failed to decode prompt %q: %w", p.ID, err)
		}
		data.Prompts[p.ID] = config
	}
	return nil
}

// decodeSourceConfig decodes a Source entity's config JSON into a SourceConfig.
func decodeSourceConfig(ctx context.Context, src *ent.Source) (sources.SourceConfig, error) {
	decoder, err := util.NewStrictDecoder(src.Config)
	if err != nil {
		return nil, err
	}
	return sources.DecodeConfig(ctx, src.Kind, src.ID, decoder)
}

// decodeAuthServiceConfig decodes an AuthService entity's config JSON into an AuthServiceConfig.
func decodeAuthServiceConfig(_ context.Context, as *ent.AuthService) (auth.AuthServiceConfig, error) {
	// Handle auth service based on kind
	switch as.Kind {
	case google.AuthServiceKind:
		cfg := google.Config{
			Name: as.ID,
			Kind: as.Kind,
		}
		// Extract clientId from config
		if clientID, ok := as.Config["clientId"].(string); ok {
			cfg.ClientID = clientID
		}
		return cfg, nil
	default:
		return nil, fmt.Errorf("unknown auth service kind: %q", as.Kind)
	}
}

// decodeToolConfig decodes a Tool entity's config JSON into a ToolConfig.
func decodeToolConfig(ctx context.Context, t *ent.Tool) (tools.ToolConfig, error) {
	decoder, err := util.NewStrictDecoder(t.Config)
	if err != nil {
		return nil, err
	}
	return tools.DecodeConfig(ctx, t.Kind, t.ID, decoder)
}

// decodePromptConfig decodes a Prompt entity's config JSON into a PromptConfig.
func decodePromptConfig(ctx context.Context, p *ent.Prompt) (prompts.PromptConfig, error) {
	decoder, err := util.NewStrictDecoder(p.Config)
	if err != nil {
		return nil, err
	}
	return prompts.DecodeConfig(ctx, p.Kind, p.ID, decoder)
}

// SaveSource saves a source configuration to the database.
func (s *Store) SaveSource(ctx context.Context, name, kind string, config map[string]any) error {
	// Try to update first, if not exists then create
	exists, err := s.client.Source.Query().Where().Count(ctx)
	if err != nil {
		return err
	}

	if exists > 0 {
		// Check if this specific source exists
		_, err := s.client.Source.Get(ctx, name)
		if err == nil {
			// Update existing
			return s.client.Source.UpdateOneID(name).
				SetKind(kind).
				SetConfig(config).
				Exec(ctx)
		}
	}

	// Create new
	return s.client.Source.Create().
		SetID(name).
		SetKind(kind).
		SetConfig(config).
		Exec(ctx)
}

// SaveTool saves a tool configuration to the database.
func (s *Store) SaveTool(ctx context.Context, name, kind string, sourceName *string, config map[string]any) error {
	_, err := s.client.Tool.Get(ctx, name)
	if err == nil {
		// Update existing
		update := s.client.Tool.UpdateOneID(name).
			SetKind(kind).
			SetConfig(config)
		if sourceName != nil {
			update = update.SetSourceName(*sourceName)
		} else {
			update = update.ClearSourceName()
		}
		return update.Exec(ctx)
	}

	// Create new
	create := s.client.Tool.Create().
		SetID(name).
		SetKind(kind).
		SetConfig(config)
	if sourceName != nil {
		create = create.SetSourceName(*sourceName)
	}
	return create.Exec(ctx)
}

// SaveToolset saves a toolset configuration to the database.
func (s *Store) SaveToolset(ctx context.Context, name string, toolNames []string) error {
	_, err := s.client.Toolset.Get(ctx, name)
	if err == nil {
		// Update existing
		return s.client.Toolset.UpdateOneID(name).
			SetToolNames(toolNames).
			Exec(ctx)
	}

	// Create new
	return s.client.Toolset.Create().
		SetID(name).
		SetToolNames(toolNames).
		Exec(ctx)
}

// SavePrompt saves a prompt configuration to the database.
func (s *Store) SavePrompt(ctx context.Context, name, kind string, config map[string]any) error {
	_, err := s.client.Prompt.Get(ctx, name)
	if err == nil {
		// Update existing
		return s.client.Prompt.UpdateOneID(name).
			SetKind(kind).
			SetConfig(config).
			Exec(ctx)
	}

	// Create new
	return s.client.Prompt.Create().
		SetID(name).
		SetKind(kind).
		SetConfig(config).
		Exec(ctx)
}

// SaveAuthService saves an auth service configuration to the database.
func (s *Store) SaveAuthService(ctx context.Context, name, kind string, config map[string]any) error {
	_, err := s.client.AuthService.Get(ctx, name)
	if err == nil {
		// Update existing
		return s.client.AuthService.UpdateOneID(name).
			SetKind(kind).
			SetConfig(config).
			Exec(ctx)
	}

	// Create new
	return s.client.AuthService.Create().
		SetID(name).
		SetKind(kind).
		SetConfig(config).
		Exec(ctx)
}

// SavePromptset saves a promptset configuration to the database.
func (s *Store) SavePromptset(ctx context.Context, name string, promptNames []string) error {
	_, err := s.client.Promptset.Get(ctx, name)
	if err == nil {
		// Update existing
		return s.client.Promptset.UpdateOneID(name).
			SetPromptNames(promptNames).
			Exec(ctx)
	}

	// Create new
	return s.client.Promptset.Create().
		SetID(name).
		SetPromptNames(promptNames).
		Exec(ctx)
}

