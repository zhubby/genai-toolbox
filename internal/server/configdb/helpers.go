package configdb

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

func dbPathFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if v := r.URL.Query().Get("dbPath"); v != "" {
		return v
	}
	// 兼容常见 snake_case 传参
	return r.URL.Query().Get("db_path")
}

func openForRead(ctx context.Context, dbPath string) (*storage.Store, bool, error) {
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

func openForWrite(ctx context.Context, dbPath string) (*storage.Store, error) {
	return storage.Open(ctx, dbPath)
}

func decodeJSON(r *http.Request, v any) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	return util.DecodeJSON(r.Body, v)
}

func normalizeNumbers(m map[string]any) (map[string]any, error) {
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

func writeError(deps Dependencies, w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	status := http.StatusInternalServerError
	switch {
	case ent.IsNotFound(err):
		status = http.StatusNotFound
	case ent.IsConstraintError(err):
		status = http.StatusConflict
	case errors.Is(err, errBadRequest):
		status = http.StatusBadRequest
	}
	if deps.Logger != nil {
		deps.Logger.DebugContext(r.Context(), err.Error())
	}
	_ = render.Render(w, r, newErrResponse(err, status))
}

var errBadRequest = errors.New("bad request")
