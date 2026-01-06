package storage

import (
	"context"

	"github.com/googleapis/genai-toolbox/internal/storage/ent"
	"github.com/googleapis/genai-toolbox/internal/storage/ent/authservice"
	"github.com/googleapis/genai-toolbox/internal/storage/ent/prompt"
	"github.com/googleapis/genai-toolbox/internal/storage/ent/promptset"
	"github.com/googleapis/genai-toolbox/internal/storage/ent/source"
	"github.com/googleapis/genai-toolbox/internal/storage/ent/tool"
	"github.com/googleapis/genai-toolbox/internal/storage/ent/toolset"
)

// ===== Source =====

func (s *Store) CreateSource(ctx context.Context, name, kind string, config map[string]any) (*ent.Source, error) {
	return s.client.Source.Create().
		SetID(name).
		SetKind(kind).
		SetConfig(config).
		Save(ctx)
}

func (s *Store) UpdateSource(ctx context.Context, name, kind string, config map[string]any) (*ent.Source, error) {
	return s.client.Source.UpdateOneID(name).
		SetKind(kind).
		SetConfig(config).
		Save(ctx)
}

func (s *Store) GetSource(ctx context.Context, name string) (*ent.Source, error) {
	return s.client.Source.Get(ctx, name)
}

func (s *Store) ListSources(ctx context.Context) ([]*ent.Source, error) {
	return s.client.Source.Query().
		Order(source.ByID()).
		All(ctx)
}

func (s *Store) DeleteSource(ctx context.Context, name string) error {
	return s.client.Source.DeleteOneID(name).Exec(ctx)
}

// ===== Tool =====

func (s *Store) CreateTool(ctx context.Context, name, kind string, sourceName *string, config map[string]any) (*ent.Tool, error) {
	create := s.client.Tool.Create().
		SetID(name).
		SetKind(kind).
		SetConfig(config)
	if sourceName != nil {
		create = create.SetSourceName(*sourceName)
	}
	return create.Save(ctx)
}

func (s *Store) UpdateTool(ctx context.Context, name, kind string, sourceName *string, config map[string]any) (*ent.Tool, error) {
	update := s.client.Tool.UpdateOneID(name).
		SetKind(kind).
		SetConfig(config)
	if sourceName != nil {
		update = update.SetSourceName(*sourceName)
	} else {
		update = update.ClearSourceName()
	}
	return update.Save(ctx)
}

func (s *Store) GetTool(ctx context.Context, name string) (*ent.Tool, error) {
	return s.client.Tool.Get(ctx, name)
}

func (s *Store) ListTools(ctx context.Context) ([]*ent.Tool, error) {
	return s.client.Tool.Query().
		Order(tool.ByID()).
		All(ctx)
}

func (s *Store) DeleteTool(ctx context.Context, name string) error {
	return s.client.Tool.DeleteOneID(name).Exec(ctx)
}

// ===== AuthService =====

func (s *Store) CreateAuthService(ctx context.Context, name, kind string, config map[string]any) (*ent.AuthService, error) {
	return s.client.AuthService.Create().
		SetID(name).
		SetKind(kind).
		SetConfig(config).
		Save(ctx)
}

func (s *Store) UpdateAuthService(ctx context.Context, name, kind string, config map[string]any) (*ent.AuthService, error) {
	return s.client.AuthService.UpdateOneID(name).
		SetKind(kind).
		SetConfig(config).
		Save(ctx)
}

func (s *Store) GetAuthService(ctx context.Context, name string) (*ent.AuthService, error) {
	return s.client.AuthService.Get(ctx, name)
}

func (s *Store) ListAuthServices(ctx context.Context) ([]*ent.AuthService, error) {
	return s.client.AuthService.Query().
		Order(authservice.ByID()).
		All(ctx)
}

func (s *Store) DeleteAuthService(ctx context.Context, name string) error {
	return s.client.AuthService.DeleteOneID(name).Exec(ctx)
}

// ===== Toolset =====

func (s *Store) CreateToolset(ctx context.Context, name string, toolNames []string) (*ent.Toolset, error) {
	return s.client.Toolset.Create().
		SetID(name).
		SetToolNames(toolNames).
		Save(ctx)
}

func (s *Store) UpdateToolset(ctx context.Context, name string, toolNames []string) (*ent.Toolset, error) {
	return s.client.Toolset.UpdateOneID(name).
		SetToolNames(toolNames).
		Save(ctx)
}

func (s *Store) GetToolset(ctx context.Context, name string) (*ent.Toolset, error) {
	return s.client.Toolset.Get(ctx, name)
}

func (s *Store) ListToolsets(ctx context.Context) ([]*ent.Toolset, error) {
	return s.client.Toolset.Query().
		Order(toolset.ByID()).
		All(ctx)
}

func (s *Store) DeleteToolset(ctx context.Context, name string) error {
	return s.client.Toolset.DeleteOneID(name).Exec(ctx)
}

// ===== Prompt =====

func (s *Store) CreatePrompt(ctx context.Context, name, kind string, config map[string]any) (*ent.Prompt, error) {
	return s.client.Prompt.Create().
		SetID(name).
		SetKind(kind).
		SetConfig(config).
		Save(ctx)
}

func (s *Store) UpdatePrompt(ctx context.Context, name, kind string, config map[string]any) (*ent.Prompt, error) {
	return s.client.Prompt.UpdateOneID(name).
		SetKind(kind).
		SetConfig(config).
		Save(ctx)
}

func (s *Store) GetPrompt(ctx context.Context, name string) (*ent.Prompt, error) {
	return s.client.Prompt.Get(ctx, name)
}

func (s *Store) ListPrompts(ctx context.Context) ([]*ent.Prompt, error) {
	return s.client.Prompt.Query().
		Order(prompt.ByID()).
		All(ctx)
}

func (s *Store) DeletePrompt(ctx context.Context, name string) error {
	return s.client.Prompt.DeleteOneID(name).Exec(ctx)
}

// ===== Promptset =====

func (s *Store) CreatePromptset(ctx context.Context, name string, promptNames []string) (*ent.Promptset, error) {
	return s.client.Promptset.Create().
		SetID(name).
		SetPromptNames(promptNames).
		Save(ctx)
}

func (s *Store) UpdatePromptset(ctx context.Context, name string, promptNames []string) (*ent.Promptset, error) {
	return s.client.Promptset.UpdateOneID(name).
		SetPromptNames(promptNames).
		Save(ctx)
}

func (s *Store) GetPromptset(ctx context.Context, name string) (*ent.Promptset, error) {
	return s.client.Promptset.Get(ctx, name)
}

func (s *Store) ListPromptsets(ctx context.Context) ([]*ent.Promptset, error) {
	return s.client.Promptset.Query().
		Order(promptset.ByID()).
		All(ctx)
}

func (s *Store) DeletePromptset(ctx context.Context, name string) error {
	return s.client.Promptset.DeleteOneID(name).Exec(ctx)
}
