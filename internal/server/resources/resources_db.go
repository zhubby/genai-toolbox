package resources

import (
	"context"

	"github.com/googleapis/genai-toolbox/internal/auth"
	"github.com/googleapis/genai-toolbox/internal/auth/google"
	"github.com/googleapis/genai-toolbox/internal/log"
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
	logger        log.Logger
	serverVersion string
}

func NewResourceManagerForDB(dbPath string, tracer trace.Tracer, logger log.Logger, serverVersion string) *ResourceManagerForDB {
	return &ResourceManagerForDB{
		dbPath:        dbPath,
		tracer:        tracer,
		logger:        logger,
		serverVersion: serverVersion,
	}
}

func (r *ResourceManagerForDB) baseCtx() context.Context {
	ctx := context.Background()
	// Some config decoding paths (e.g. parameters) emit warnings via logger from context.
	// Attach it when available to avoid decode failures and keep warnings.
	if r.logger != nil {
		ctx = util.WithLogger(ctx, r.logger)
	}
	return ctx
}

func (r *ResourceManagerForDB) warn(ctx context.Context, format string, args ...any) {
	if r.logger == nil {
		return
	}
	r.logger.WarnContext(ctx, format, args...)
}

func (r *ResourceManagerForDB) debug(ctx context.Context, format string, args ...any) {
	if r.logger == nil {
		return
	}
	r.logger.DebugContext(ctx, format, args...)
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
	ctx := r.baseCtx()
	store, ok, err := r.openForRead(ctx)
	if err != nil {
		r.warn(ctx, "open config db failed (GetSource %q): %v", sourceName, err)
		return nil, false
	}
	if !ok {
		return nil, false
	}
	defer store.Close() //nolint:errcheck

	row, err := store.GetSource(ctx, sourceName)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false
		}
		r.warn(ctx, "failed to load source %q from config db: %v", sourceName, err)
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
		r.warn(ctx, "failed to create decoder for source %q: %v", sourceName, err)
		return nil, false
	}
	cfg, err := sources.DecodeConfig(ctx, row.Kind, row.ID, dec)
	if err != nil {
		r.warn(ctx, "failed to decode source %q (kind=%q): %v", sourceName, row.Kind, err)
		return nil, false
	}
	src, err := cfg.Initialize(ctx, r.tracer)
	if err != nil {
		r.warn(ctx, "failed to initialize source %q (kind=%q): %v", sourceName, row.Kind, err)
		return nil, false
	}
	return src, true
}

func (r *ResourceManagerForDB) GetAuthService(authServiceName string) (auth.AuthService, bool) {
	ctx := r.baseCtx()
	store, ok, err := r.openForRead(ctx)
	if err != nil {
		r.warn(ctx, "open config db failed (GetAuthService %q): %v", authServiceName, err)
		return nil, false
	}
	if !ok {
		return nil, false
	}
	defer store.Close() //nolint:errcheck

	row, err := store.GetAuthService(ctx, authServiceName)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false
		}
		r.warn(ctx, "failed to load auth service %q from config db: %v", authServiceName, err)
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
		r.warn(ctx, "failed to create decoder for auth service %q: %v", authServiceName, err)
		return nil, false
	}
	switch row.Kind {
	case google.AuthServiceKind:
		cfg := google.Config{Name: row.ID}
		if err := dec.Decode(&cfg); err != nil {
			r.warn(ctx, "failed to decode auth service %q (kind=%q): %v", authServiceName, row.Kind, err)
			return nil, false
		}
		a, err := cfg.Initialize()
		if err != nil {
			r.warn(ctx, "failed to initialize auth service %q (kind=%q): %v", authServiceName, row.Kind, err)
			return nil, false
		}
		return a, true
	default:
		r.debug(ctx, "unknown auth service kind %q for %q", row.Kind, authServiceName)
		return nil, false
	}
}

func (r *ResourceManagerForDB) GetTool(toolName string) (tools.Tool, bool) {
	ctx := r.baseCtx()
	store, ok, err := r.openForRead(ctx)
	if err != nil {
		r.warn(ctx, "open config db failed (GetTool %q): %v", toolName, err)
		return nil, false
	}
	if !ok {
		return nil, false
	}
	defer store.Close() //nolint:errcheck

	row, err := store.GetTool(ctx, toolName)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false
		}
		r.warn(ctx, "failed to load tool %q from config db: %v", toolName, err)
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
		r.warn(ctx, "failed to create decoder for tool %q: %v", toolName, err)
		return nil, false
	}
	cfg, err := tools.DecodeConfig(ctx, row.Kind, row.ID, dec)
	if err != nil {
		r.warn(ctx, "failed to decode tool %q (kind=%q): %v", toolName, row.Kind, err)
		return nil, false
	}

	srcs := make(map[string]sources.Source)
	if row.SourceName != nil {
		src, ok := r.GetSource(*row.SourceName)
		if !ok {
			r.warn(ctx, "tool %q references missing source %q", toolName, *row.SourceName)
			return nil, false
		}
		srcs[*row.SourceName] = src
	}
	t, err := cfg.Initialize(srcs)
	if err != nil {
		r.warn(ctx, "failed to initialize tool %q (kind=%q): %v", toolName, row.Kind, err)
		return nil, false
	}
	return t, true
}

func (r *ResourceManagerForDB) GetToolset(toolsetName string) (tools.Toolset, bool) {
	ctx := r.baseCtx()
	store, ok, err := r.openForRead(ctx)
	if err != nil {
		r.warn(ctx, "open config db failed (GetToolset %q): %v", toolsetName, err)
		return tools.Toolset{}, false
	}
	if !ok {
		return tools.Toolset{}, false
	}
	defer store.Close() //nolint:errcheck

	var toolNames []string
	if toolsetName == "" {
		// Default toolset: include all tools currently in DB.
		rows, err := store.ListTools(ctx)
		if err != nil {
			r.warn(ctx, "failed to list tools for default toolset: %v", err)
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
			r.warn(ctx, "failed to load toolset %q from config db: %v", toolsetName, err)
			return tools.Toolset{}, false
		}
		toolNames = row.ToolNames
	}

	toolsMap := make(map[string]tools.Tool, len(toolNames))
	for _, name := range toolNames {
		t, ok := r.GetTool(name)
		if !ok {
			r.warn(ctx, "toolset %q references missing tool %q", toolsetName, name)
			return tools.Toolset{}, false
		}
		toolsMap[name] = t
	}

	cfg := tools.ToolsetConfig{Name: toolsetName, ToolNames: toolNames}
	ts, err := cfg.Initialize(r.serverVersion, toolsMap)
	if err != nil {
		r.warn(ctx, "failed to initialize toolset %q: %v", toolsetName, err)
		return tools.Toolset{}, false
	}
	return ts, true
}

func (r *ResourceManagerForDB) GetPrompt(promptName string) (prompts.Prompt, bool) {
	ctx := r.baseCtx()
	store, ok, err := r.openForRead(ctx)
	if err != nil {
		r.warn(ctx, "open config db failed (GetPrompt %q): %v", promptName, err)
		return nil, false
	}
	if !ok {
		return nil, false
	}
	defer store.Close() //nolint:errcheck

	row, err := store.GetPrompt(ctx, promptName)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false
		}
		r.warn(ctx, "failed to load prompt %q from config db: %v", promptName, err)
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
		r.warn(ctx, "failed to create decoder for prompt %q: %v", promptName, err)
		return nil, false
	}
	cfg, err := prompts.DecodeConfig(ctx, row.Kind, row.ID, dec)
	if err != nil {
		r.warn(ctx, "failed to decode prompt %q (kind=%q): %v", promptName, row.Kind, err)
		return nil, false
	}
	p, err := cfg.Initialize()
	if err != nil {
		r.warn(ctx, "failed to initialize prompt %q (kind=%q): %v", promptName, row.Kind, err)
		return nil, false
	}
	return p, true
}

func (r *ResourceManagerForDB) GetPromptset(promptsetName string) (prompts.Promptset, bool) {
	ctx := r.baseCtx()
	store, ok, err := r.openForRead(ctx)
	if err != nil {
		r.warn(ctx, "open config db failed (GetPromptset %q): %v", promptsetName, err)
		return prompts.Promptset{}, false
	}
	if !ok {
		return prompts.Promptset{}, false
	}
	defer store.Close() //nolint:errcheck

	var promptNames []string
	if promptsetName == "" {
		rows, err := store.ListPrompts(ctx)
		if err != nil {
			r.warn(ctx, "failed to list prompts for default promptset: %v", err)
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
			r.warn(ctx, "failed to load promptset %q from config db: %v", promptsetName, err)
			return prompts.Promptset{}, false
		}
		promptNames = row.PromptNames
	}

	promptsMap := make(map[string]prompts.Prompt, len(promptNames))
	for _, name := range promptNames {
		p, ok := r.GetPrompt(name)
		if !ok {
			r.warn(ctx, "promptset %q references missing prompt %q", promptsetName, name)
			return prompts.Promptset{}, false
		}
		promptsMap[name] = p
	}

	cfg := prompts.PromptsetConfig{Name: promptsetName, PromptNames: promptNames}
	ps, err := cfg.Initialize(r.serverVersion, promptsMap)
	if err != nil {
		r.warn(ctx, "failed to initialize promptset %q: %v", promptsetName, err)
		return prompts.Promptset{}, false
	}
	return ps, true
}

func (r *ResourceManagerForDB) GetAuthServiceMap() map[string]auth.AuthService {
	ctx := r.baseCtx()
	store, ok, err := r.openForRead(ctx)
	if err != nil {
		r.warn(ctx, "open config db failed (GetAuthServiceMap): %v", err)
		return map[string]auth.AuthService{}
	}
	if !ok {
		return map[string]auth.AuthService{}
	}
	defer store.Close() //nolint:errcheck

	rows, err := store.ListAuthServices(ctx)
	if err != nil {
		r.warn(ctx, "failed to list auth services from config db: %v", err)
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
	ctx := r.baseCtx()
	store, ok, err := r.openForRead(ctx)
	if err != nil {
		r.warn(ctx, "open config db failed (GetToolsMap): %v", err)
		return map[string]tools.Tool{}
	}
	if !ok {
		return map[string]tools.Tool{}
	}
	defer store.Close() //nolint:errcheck

	rows, err := store.ListTools(ctx)
	if err != nil {
		r.warn(ctx, "failed to list tools from config db: %v", err)
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
	ctx := r.baseCtx()
	store, ok, err := r.openForRead(ctx)
	if err != nil {
		r.warn(ctx, "open config db failed (GetPromptsMap): %v", err)
		return map[string]prompts.Prompt{}
	}
	if !ok {
		return map[string]prompts.Prompt{}
	}
	defer store.Close() //nolint:errcheck

	rows, err := store.ListPrompts(ctx)
	if err != nil {
		r.warn(ctx, "failed to list prompts from config db: %v", err)
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


