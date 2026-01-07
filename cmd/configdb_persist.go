package cmd

import (
	"context"
	"fmt"

	yaml "github.com/goccy/go-yaml"
	"github.com/googleapis/genai-toolbox/internal/storage"
	"github.com/googleapis/genai-toolbox/internal/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func saveMergedConfigToDB(ctx context.Context, dbPath string, cfg *storage.ConfigData) error {
	if cfg == nil {
		return nil
	}

	var tracer trace.Tracer
	if inst, err := util.InstrumentationFromContext(ctx); err == nil && inst != nil {
		tracer = inst.Tracer
	}
	if tracer == nil {
		tracer = trace.NewNoopTracerProvider().Tracer("noop")
	}

	ctx, span := tracer.Start(ctx, "toolbox/cmd/configdb/persist")
	defer span.End()
	span.SetAttributes(attribute.String("db_path", dbPath))

	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	defer store.Close() //nolint:errcheck

	// Sources
	for name, sc := range cfg.SourceConfigs {
		m, err := toMap(sc)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("marshal source %q: %w", name, err)
		}
		kind, _ := m["kind"].(string)
		if kind == "" {
			kind = sc.SourceConfigKind()
			m["kind"] = kind
		}
		m["name"] = name
		if err := store.SaveSource(ctx, name, kind, m); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("save source %q: %w", name, err)
		}
	}

	// AuthServices
	for name, ac := range cfg.AuthServiceConfigs {
		m, err := toMap(ac)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("marshal authservice %q: %w", name, err)
		}
		kind, _ := m["kind"].(string)
		if kind == "" {
			kind = ac.AuthServiceConfigKind()
			m["kind"] = kind
		}
		m["name"] = name
		if err := store.SaveAuthService(ctx, name, kind, m); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("save authservice %q: %w", name, err)
		}
	}

	// Tools
	for name, tc := range cfg.ToolConfigs {
		m, err := toMap(tc)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("marshal tool %q: %w", name, err)
		}
		kind, _ := m["kind"].(string)
		if kind == "" {
			kind = tc.ToolConfigKind()
			m["kind"] = kind
		}
		m["name"] = name

		// If tool declares source in config, keep sourceName column aligned.
		var sourceName *string
		if src, ok := m["source"].(string); ok && src != "" {
			sourceName = &src
		}
		if err := store.SaveTool(ctx, name, kind, sourceName, m); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("save tool %q: %w", name, err)
		}
	}

	// Toolsets
	for name, ts := range cfg.ToolsetConfigs {
		if err := store.SaveToolset(ctx, name, ts.ToolNames); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("save toolset %q: %w", name, err)
		}
	}

	// Prompts
	for name, pc := range cfg.PromptConfigs {
		m, err := toMap(pc)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("marshal prompt %q: %w", name, err)
		}
		kind, _ := m["kind"].(string)
		if kind == "" {
			kind = pc.PromptConfigKind()
			m["kind"] = kind
		}
		m["name"] = name
		if err := store.SavePrompt(ctx, name, kind, m); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("save prompt %q: %w", name, err)
		}
	}

	return nil
}

func toMap(v any) (map[string]any, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}


