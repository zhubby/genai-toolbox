package resources

import (
	"context"

	"github.com/googleapis/genai-toolbox/internal/auth"
	"github.com/googleapis/genai-toolbox/internal/auth/google"
	"github.com/googleapis/genai-toolbox/internal/prompts"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"github.com/googleapis/genai-toolbox/internal/storage"
	"github.com/googleapis/genai-toolbox/internal/storage/ent"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"github.com/googleapis/genai-toolbox/internal/util"
	"go.opentelemetry.io/otel/trace"
)

// ResourceManagerForDB is a DB-backed implementation of Manager.
// It loads configurations from the SQLite config DB on demand, so changes made via
// /api/config are immediately reflected in runtime calls without folder watching.
//
// NOTE: This implementation does not keep an in-memory cache of initialized resources.
// Each call may initialize new source connections and tool objects.
type ResourceManagerForDB struct {
	dbPath        string
	tracer        trace.Tracer
	serverVersion string
}

func NewResourceManagerForDB(dbPath string, tracer trace.Tracer, serverVersion string) *ResourceManagerForDB {
	return &ResourceManagerForDB{
		dbPath:        dbPath,
		tracer:        tracer,
		serverVersion: serverVersion,
	}
}

func (r *ResourceManagerForDB) openForRead(ctx context.Context) (*storage.Store, bool, error) {
	exists, err := storage.Exists(r.dbPath)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}
	s, err := storage.Open(ctx, r.dbPath)
	if err != nil {
		return nil, false, err
	}
	return s, true, nil
}

func (r *ResourceManagerForDB) GetSource(sourceName string) (sources.Source, bool) {
	ctx := context.Background()
	store, ok, err := r.openForRead(ctx)
	if err != nil || !ok {
		return nil, false
	}
	defer store.Close() //nolint:errcheck

	row, err := store.GetSource(ctx, sourceName)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false
		}
		return nil, false
	}

	full := map[string]any{
		"kind": row.Kind,
		"name": row.ID,
	}
	for k, v := range row.Config {
		full[k] = v
	}
	dec, err := util.NewStrictDecoder(full)
	if err != nil {
		return nil, false
	}
	cfg, err := sources.DecodeConfig(ctx, row.Kind, row.ID, dec)
	if err != nil {
		return nil, false
	}
	src, err := cfg.Initialize(ctx, r.tracer)
	if err != nil {
		return nil, false
	}
	return src, true
}

func (r *ResourceManagerForDB) GetAuthService(authServiceName string) (auth.AuthService, bool) {
	ctx := context.Background()
	store, ok, err := r.openForRead(ctx)
	if err != nil || !ok {
		return nil, false
	}
	defer store.Close() //nolint:errcheck

	row, err := store.GetAuthService(ctx, authServiceName)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false
		}
		return nil, false
	}

	full := map[string]any{
		"kind": row.Kind,
		"name": row.ID,
	}
	for k, v := range row.Config {
		full[k] = v
	}
	dec, err := util.NewStrictDecoder(full)
	if err != nil {
		return nil, false
	}
	switch row.Kind {
	case google.AuthServiceKind:
		cfg := google.Config{Name: row.ID}
		if err := dec.Decode(&cfg); err != nil {
			return nil, false
		}
		a, err := cfg.Initialize()
		if err != nil {
			return nil, false
		}
		return a, true
	default:
		return nil, false
	}
}

func (r *ResourceManagerForDB) GetTool(toolName string) (tools.Tool, bool) {
	ctx := context.Background()
	store, ok, err := r.openForRead(ctx)
	if err != nil || !ok {
		return nil, false
	}
	defer store.Close() //nolint:errcheck

	row, err := store.GetTool(ctx, toolName)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false
		}
		return nil, false
	}

	// Normalize config similar to server.ToolConfigs unmarshal rules.
	full := map[string]any{
		"kind": row.Kind,
		"name": row.ID,
	}
	for k, v := range row.Config {
		full[k] = v
	}
	if full["authRequired"] == nil {
		full["authRequired"] = []string{}
	}
	if full["authRequired"] != nil && full["useClientOAuth"] == true {
		return nil, false
	}
	// Ensure source field is present for tool configs that require it.
	if row.SourceName != nil {
		if _, ok := full["source"]; !ok {
			full["source"] = *row.SourceName
		}
	}

	dec, err := util.NewStrictDecoder(full)
	if err != nil {
		return nil, false
	}
	cfg, err := tools.DecodeConfig(ctx, row.Kind, row.ID, dec)
	if err != nil {
		return nil, false
	}

	srcs := make(map[string]sources.Source)
	if row.SourceName != nil {
		src, ok := r.GetSource(*row.SourceName)
		if !ok {
			return nil, false
		}
		srcs[*row.SourceName] = src
	}
	t, err := cfg.Initialize(srcs)
	if err != nil {
		return nil, false
	}
	return t, true
}

func (r *ResourceManagerForDB) GetToolset(toolsetName string) (tools.Toolset, bool) {
	ctx := context.Background()
	store, ok, err := r.openForRead(ctx)
	if err != nil || !ok {
		return tools.Toolset{}, false
	}
	defer store.Close() //nolint:errcheck

	var toolNames []string
	if toolsetName == "" {
		// Default toolset: include all tools currently in DB.
		rows, err := store.ListTools(ctx)
		if err != nil {
			return tools.Toolset{}, false
		}
		for _, t := range rows {
			toolNames = append(toolNames, t.ID)
		}
	} else {
		row, err := store.GetToolset(ctx, toolsetName)
		if err != nil {
			if ent.IsNotFound(err) {
				return tools.Toolset{}, false
			}
			return tools.Toolset{}, false
		}
		toolNames = row.ToolNames
	}

	toolsMap := make(map[string]tools.Tool, len(toolNames))
	for _, name := range toolNames {
		t, ok := r.GetTool(name)
		if !ok {
			return tools.Toolset{}, false
		}
		toolsMap[name] = t
	}

	cfg := tools.ToolsetConfig{Name: toolsetName, ToolNames: toolNames}
	ts, err := cfg.Initialize(r.serverVersion, toolsMap)
	if err != nil {
		return tools.Toolset{}, false
	}
	return ts, true
}

func (r *ResourceManagerForDB) GetPrompt(promptName string) (prompts.Prompt, bool) {
	ctx := context.Background()
	store, ok, err := r.openForRead(ctx)
	if err != nil || !ok {
		return nil, false
	}
	defer store.Close() //nolint:errcheck

	row, err := store.GetPrompt(ctx, promptName)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false
		}
		return nil, false
	}

	full := map[string]any{
		"kind": row.Kind,
		"name": row.ID,
	}
	for k, v := range row.Config {
		full[k] = v
	}
	dec, err := util.NewStrictDecoder(full)
	if err != nil {
		return nil, false
	}
	cfg, err := prompts.DecodeConfig(ctx, row.Kind, row.ID, dec)
	if err != nil {
		return nil, false
	}
	p, err := cfg.Initialize()
	if err != nil {
		return nil, false
	}
	return p, true
}

func (r *ResourceManagerForDB) GetPromptset(promptsetName string) (prompts.Promptset, bool) {
	ctx := context.Background()
	store, ok, err := r.openForRead(ctx)
	if err != nil || !ok {
		return prompts.Promptset{}, false
	}
	defer store.Close() //nolint:errcheck

	var promptNames []string
	if promptsetName == "" {
		rows, err := store.ListPrompts(ctx)
		if err != nil {
			return prompts.Promptset{}, false
		}
		for _, p := range rows {
			promptNames = append(promptNames, p.ID)
		}
	} else {
		row, err := store.GetPromptset(ctx, promptsetName)
		if err != nil {
			if ent.IsNotFound(err) {
				return prompts.Promptset{}, false
			}
			return prompts.Promptset{}, false
		}
		promptNames = row.PromptNames
	}

	promptsMap := make(map[string]prompts.Prompt, len(promptNames))
	for _, name := range promptNames {
		p, ok := r.GetPrompt(name)
		if !ok {
			return prompts.Promptset{}, false
		}
		promptsMap[name] = p
	}

	cfg := prompts.PromptsetConfig{Name: promptsetName, PromptNames: promptNames}
	ps, err := cfg.Initialize(r.serverVersion, promptsMap)
	if err != nil {
		return prompts.Promptset{}, false
	}
	return ps, true
}

func (r *ResourceManagerForDB) GetAuthServiceMap() map[string]auth.AuthService {
	ctx := context.Background()
	store, ok, err := r.openForRead(ctx)
	if err != nil || !ok {
		return map[string]auth.AuthService{}
	}
	defer store.Close() //nolint:errcheck

	rows, err := store.ListAuthServices(ctx)
	if err != nil {
		return map[string]auth.AuthService{}
	}
	out := make(map[string]auth.AuthService, len(rows))
	for _, row := range rows {
		a, ok := r.GetAuthService(row.ID)
		if ok {
			out[row.ID] = a
		}
	}
	return out
}

func (r *ResourceManagerForDB) GetToolsMap() map[string]tools.Tool {
	ctx := context.Background()
	store, ok, err := r.openForRead(ctx)
	if err != nil || !ok {
		return map[string]tools.Tool{}
	}
	defer store.Close() //nolint:errcheck

	rows, err := store.ListTools(ctx)
	if err != nil {
		return map[string]tools.Tool{}
	}
	out := make(map[string]tools.Tool, len(rows))
	for _, row := range rows {
		t, ok := r.GetTool(row.ID)
		if ok {
			out[row.ID] = t
		}
	}
	return out
}

func (r *ResourceManagerForDB) GetPromptsMap() map[string]prompts.Prompt {
	ctx := context.Background()
	store, ok, err := r.openForRead(ctx)
	if err != nil || !ok {
		return map[string]prompts.Prompt{}
	}
	defer store.Close() //nolint:errcheck

	rows, err := store.ListPrompts(ctx)
	if err != nil {
		return map[string]prompts.Prompt{}
	}
	out := make(map[string]prompts.Prompt, len(rows))
	for _, row := range rows {
		p, ok := r.GetPrompt(row.ID)
		if ok {
			out[row.ID] = p
		}
	}
	return out
}

var _ Manager = (*ResourceManagerForDB)(nil)


