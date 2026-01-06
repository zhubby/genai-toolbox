package configdb

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/render"
	"github.com/googleapis/genai-toolbox/internal/storage"
	"github.com/googleapis/genai-toolbox/internal/storage/ent"
	"github.com/googleapis/genai-toolbox/internal/util"
)

var (
	storeMu     sync.Mutex
	storeByPath = map[string]*storage.Store{}
)

func canonicalDBPath(dbPath string) string {
	if dbPath != "" {
		return dbPath
	}
	// best-effort: match storage.Open default behavior for stable cache key
	if p, err := storage.DefaultDBPath(); err == nil {
		return p
	}
	return dbPath
}

func getOrOpenStore(ctx context.Context, dbPath string) (*storage.Store, error) {
	key := canonicalDBPath(dbPath)
	storeMu.Lock()
	if s := storeByPath[key]; s != nil {
		storeMu.Unlock()
		return s, nil
	}
	storeMu.Unlock()

	// Open may create the DB and run schema migration. It is safe to call concurrently
	// because storage.Open serializes schema migration. We still cache the resulting store.
	s, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}

	storeMu.Lock()
	// If another goroutine populated the cache meanwhile, keep the first one and close ours.
	if existing := storeByPath[key]; existing != nil {
		storeMu.Unlock()
		_ = s.Close()
		return existing, nil
	}
	storeByPath[key] = s
	storeMu.Unlock()
	return s, nil
}

func openForRead(ctx context.Context, dbPath string) (*storage.Store, bool, error) {
	exists, err := storage.Exists(dbPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check db existence: %w", err)
	}
	if !exists {
		return nil, false, nil
	}
	store, err := getOrOpenStore(ctx, dbPath)
	if err != nil {
		return nil, false, err
	}
	return store, true, nil
}

func openForWrite(ctx context.Context, dbPath string) (*storage.Store, error) {
	return getOrOpenStore(ctx, dbPath)
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
