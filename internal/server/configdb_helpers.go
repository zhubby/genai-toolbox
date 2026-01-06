package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/googleapis/genai-toolbox/internal/storage"
	"github.com/googleapis/genai-toolbox/internal/storage/ent"
	"github.com/googleapis/genai-toolbox/internal/util"
)

func configDBPathFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if v := r.URL.Query().Get("dbPath"); v != "" {
		return v
	}
	// 兼容常见 snake_case 传参
	return r.URL.Query().Get("db_path")
}

func configDBOpenForRead(ctx context.Context, dbPath string) (*storage.Store, bool, error) {
	exists, err := storage.Exists(dbPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check db existence: %w", err)
	}
	if !exists {
		return nil, false, nil
	}
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, false, err
	}
	return store, true, nil
}

func configDBOpenForWrite(ctx context.Context, dbPath string) (*storage.Store, error) {
	return storage.Open(ctx, dbPath)
}

func configDBDecodeJSON(r *http.Request, v any) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	if err := util.DecodeJSON(r.Body, v); err != nil {
		return err
	}
	return nil
}

func configDBNormalizeNumbers(m map[string]any) (map[string]any, error) {
	if m == nil {
		return map[string]any{}, nil
	}
	converted, err := util.ConvertNumbers(m)
	if err != nil {
		return nil, err
	}
	convertedMap, ok := converted.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("config is not a JSON object")
	}
	return convertedMap, nil
}

func configDBWriteError(s *Server, w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	status := http.StatusInternalServerError
	switch {
	case ent.IsNotFound(err):
		status = http.StatusNotFound
	case ent.IsConstraintError(err):
		status = http.StatusConflict
	case errors.Is(err, errConfigDBBadRequest):
		status = http.StatusBadRequest
	}
	if s != nil && s.logger != nil {
		s.logger.DebugContext(r.Context(), err.Error())
	}
	_ = render.Render(w, r, newErrResponse(err, status))
}

var errConfigDBBadRequest = errors.New("bad request")

