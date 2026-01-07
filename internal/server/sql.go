package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/googleapis/genai-toolbox/internal/util"
	"github.com/googleapis/genai-toolbox/internal/util/orderedmap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type executeSQLRequest struct {
	Source     string `json:"source"`
	Statement  string `json:"statement"`
	Parameters any    `json:"parameters,omitempty"` // []any (positional) or map[string]any (named)
	ReadOnly   *bool  `json:"readOnly,omitempty"`   // for sources that need an explicit read-only hint (e.g. Spanner)
}

type executeSQLResponse struct {
	Columns    []string `json:"columns"`
	Rows       [][]any  `json:"rows"`
	RowCount   int      `json:"rowCount"`
	DurationMs int64    `json:"durationMs"`
}

type runSQLPositional interface {
	RunSQL(ctx context.Context, statement string, params []any) (any, error)
}

type runSQLNamedWithReadOnly interface {
	RunSQL(ctx context.Context, readOnly bool, statement string, params map[string]any) (any, error)
}

func sqlExecuteHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/sql/execute")
	defer span.End()
	r = r.WithContext(ctx)

	start := time.Now()

	var req executeSQLRequest
	if err := util.DecodeJSON(r.Body, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		_ = render.Render(w, r, newErrResponse(fmt.Errorf("request body was invalid JSON: %w", err), http.StatusBadRequest))
		return
	}
	req.Source = strings.TrimSpace(req.Source)
	req.Statement = strings.TrimSpace(req.Statement)
	span.SetAttributes(attribute.String("source_name", req.Source))

	if req.Source == "" || req.Statement == "" {
		span.SetStatus(codes.Error, "missing source or statement")
		_ = render.Render(w, r, newErrResponse(fmt.Errorf("source and statement are required"), http.StatusBadRequest))
		return
	}

	src, ok := s.ResourceMgr.GetSource(req.Source)
	if !ok {
		span.SetStatus(codes.Error, "source not found")
		_ = render.Render(w, r, newErrResponse(fmt.Errorf("source %q does not exist", req.Source), http.StatusNotFound))
		return
	}

	var (
		raw any
		err error
	)

	switch typed := src.(type) {
	case runSQLPositional:
		params, perr := parsePositionalParams(req.Parameters)
		if perr != nil {
			span.SetStatus(codes.Error, perr.Error())
			_ = render.Render(w, r, newErrResponse(perr, http.StatusBadRequest))
			return
		}
		raw, err = typed.RunSQL(ctx, req.Statement, params)
	case runSQLNamedWithReadOnly:
		params, perr := parseNamedParams(req.Parameters)
		if perr != nil {
			span.SetStatus(codes.Error, perr.Error())
			_ = render.Render(w, r, newErrResponse(perr, http.StatusBadRequest))
			return
		}
		readOnly := isReadOnlyStatement(req.Statement)
		if req.ReadOnly != nil {
			readOnly = *req.ReadOnly
		}
		raw, err = typed.RunSQL(ctx, readOnly, req.Statement, params)
	default:
		span.SetStatus(codes.Error, "unsupported source")
		_ = render.Render(w, r, newErrResponse(
			fmt.Errorf("source %q (kind=%q) does not support SQL execution via this endpoint", req.Source, src.SourceKind()),
			http.StatusBadRequest,
		))
		return
	}

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		_ = render.Render(w, r, newErrResponse(fmt.Errorf("error executing SQL: %w", err), http.StatusBadRequest))
		return
	}

	cols, rows := formatTabular(raw)
	resp := executeSQLResponse{
		Columns:    cols,
		Rows:       rows,
		RowCount:   len(rows),
		DurationMs: time.Since(start).Milliseconds(),
	}
	render.JSON(w, r, resp)
}

func parsePositionalParams(v any) ([]any, error) {
	if v == nil {
		return nil, nil
	}
	switch vv := v.(type) {
	case []any:
		converted, err := util.ConvertNumbers(vv)
		if err != nil {
			return nil, fmt.Errorf("invalid parameters: %w", err)
		}
		out, ok := converted.([]any)
		if !ok {
			return nil, fmt.Errorf("parameters must be a JSON array")
		}
		return out, nil
	case map[string]any:
		return nil, fmt.Errorf("parameters must be a JSON array for positional-parameter sources")
	default:
		return nil, fmt.Errorf("parameters must be a JSON array")
	}
}

func parseNamedParams(v any) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}
	switch vv := v.(type) {
	case map[string]any:
		converted, err := util.ConvertNumbers(vv)
		if err != nil {
			return nil, fmt.Errorf("invalid parameters: %w", err)
		}
		out, ok := converted.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("parameters must be a JSON object")
		}
		return out, nil
	case []any:
		return nil, fmt.Errorf("parameters must be a JSON object for named-parameter sources")
	default:
		return nil, fmt.Errorf("parameters must be a JSON object")
	}
}

func isReadOnlyStatement(stmt string) bool {
	s := strings.TrimSpace(strings.ToLower(stmt))
	switch {
	case strings.HasPrefix(s, "select"):
		return true
	case strings.HasPrefix(s, "with"):
		return true
	case strings.HasPrefix(s, "show"):
		return true
	case strings.HasPrefix(s, "describe"):
		return true
	case strings.HasPrefix(s, "explain"):
		return true
	default:
		return false
	}
}

func formatTabular(raw any) ([]string, [][]any) {
	if raw == nil {
		return nil, nil
	}

	// Common path: most SQL sources return []any where each entry is orderedmap.Row.
	if list, ok := raw.([]any); ok {
		if len(list) == 0 {
			return nil, nil
		}

		switch first := list[0].(type) {
		case orderedmap.Row:
			cols := make([]string, 0, len(first.Columns))
			for _, c := range first.Columns {
				cols = append(cols, c.Name)
			}
			rows := make([][]any, 0, len(list))
			for _, item := range list {
				r, ok := item.(orderedmap.Row)
				if !ok {
					continue
				}
				row := make([]any, 0, len(cols))
				for _, c := range r.Columns {
					row = append(row, c.Value)
				}
				rows = append(rows, row)
			}
			return cols, rows
		case map[string]any:
			cols := make([]string, 0, len(first))
			for k := range first {
				cols = append(cols, k)
			}
			// Keep stable ordering for a given response.
			sortStrings(cols)
			rows := make([][]any, 0, len(list))
			for _, item := range list {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				row := make([]any, 0, len(cols))
				for _, c := range cols {
					row = append(row, m[c])
				}
				rows = append(rows, row)
			}
			return cols, rows
		default:
			// Fallback: treat as 1-column result.
			cols := []string{"result"}
			rows := make([][]any, 0, len(list))
			for _, item := range list {
				rows = append(rows, []any{item})
			}
			return cols, rows
		}
	}

	// Single row object.
	if m, ok := raw.(map[string]any); ok {
		cols := make([]string, 0, len(m))
		for k := range m {
			cols = append(cols, k)
		}
		sortStrings(cols)
		row := make([]any, 0, len(cols))
		for _, c := range cols {
			row = append(row, m[c])
		}
		return cols, [][]any{row}
	}

	// Final fallback.
	return []string{"result"}, [][]any{{raw}}
}

func sortStrings(ss []string) {
	// local insertion sort: avoid extra imports for small slices
	for i := 1; i < len(ss); i++ {
		for j := i; j > 0 && ss[j] < ss[j-1]; j-- {
			ss[j], ss[j-1] = ss[j-1], ss[j]
		}
	}
}
