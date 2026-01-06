package configdb

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func listAuthServicesHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/list")
	defer span.End()

	dbPath := deps.DBPath
	store, ok, err := openForRead(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	if !ok {
		render.JSON(w, r, []AuthServiceDTO{})
		return
	}

	rows, err := store.ListAuthServices(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	out := make([]AuthServiceDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, AuthServiceDTO{
			Name:      row.ID,
			Kind:      row.Kind,
			Config:    row.Config,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	render.JSON(w, r, out)
}

func getAuthServiceHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/get")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("authservice_name", name))

	dbPath := deps.DBPath
	store, ok, err := openForRead(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	if !ok {
		err := fmt.Errorf("config db not found")
		span.SetStatus(codes.Error, err.Error())
		_ = render.Render(w, r, newErrResponse(err, http.StatusNotFound))
		return
	}

	row, err := store.GetAuthService(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	render.JSON(w, r, AuthServiceDTO{
		Name:      row.ID,
		Kind:      row.Kind,
		Config:    row.Config,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}

func createAuthServiceHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/create")
	defer span.End()

	var req CreateAuthServiceRequest
	if err := decodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, fmt.Errorf("%w: invalid JSON: %v", errBadRequest, err))
		return
	}
	if req.Name == "" || req.Kind == "" {
		err := fmt.Errorf("%w: name/kind is required", errBadRequest)
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	cfg, err := normalizeNumbers(req.Config)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, fmt.Errorf("%w: %v", errBadRequest, err))
		return
	}
	// Persist kind into config JSON to keep DB data self-contained.
	cfg["kind"] = req.Kind

	dbPath := deps.DBPath
	store, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}

	row, err := store.CreateAuthService(ctx, req.Name, req.Kind, cfg)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, AuthServiceDTO{
		Name:      row.ID,
		Kind:      row.Kind,
		Config:    row.Config,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}

func updateAuthServiceHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/update")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("authservice_name", name))

	var req UpdateAuthServiceRequest
	if err := decodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, fmt.Errorf("%w: invalid JSON: %v", errBadRequest, err))
		return
	}
	if req.Kind == "" {
		err := fmt.Errorf("%w: kind is required", errBadRequest)
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	cfg, err := normalizeNumbers(req.Config)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, fmt.Errorf("%w: %v", errBadRequest, err))
		return
	}
	// Persist kind into config JSON to keep DB data self-contained.
	cfg["kind"] = req.Kind

	dbPath := deps.DBPath
	store, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}

	row, err := store.UpdateAuthService(ctx, name, req.Kind, cfg)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	render.JSON(w, r, AuthServiceDTO{
		Name:      row.ID,
		Kind:      row.Kind,
		Config:    row.Config,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}

func deleteAuthServiceHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/delete")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("authservice_name", name))

	dbPath := deps.DBPath
	store, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}

	if err := store.DeleteAuthService(ctx, name); err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
