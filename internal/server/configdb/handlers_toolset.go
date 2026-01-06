package configdb

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func listToolsetsHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/toolset/list")
	defer span.End()

	dbPath := dbPathFromRequest(r)
	store, ok, err := openForRead(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	if !ok {
		render.JSON(w, r, []ToolsetDTO{})
		return
	}
	defer store.Close()

	rows, err := store.ListToolsets(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	out := make([]ToolsetDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, ToolsetDTO{
			Name:      row.ID,
			ToolNames: row.ToolNames,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	render.JSON(w, r, out)
}

func getToolsetHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/toolset/get")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("toolset_name", name))

	dbPath := dbPathFromRequest(r)
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
	defer store.Close()

	row, err := store.GetToolset(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	render.JSON(w, r, ToolsetDTO{
		Name:      row.ID,
		ToolNames: row.ToolNames,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}

func createToolsetHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/toolset/create")
	defer span.End()

	var req CreateToolsetRequest
	if err := decodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, fmt.Errorf("%w: invalid JSON: %v", errBadRequest, err))
		return
	}
	if req.Name == "" {
		err := fmt.Errorf("%w: name is required", errBadRequest)
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}

	dbPath := dbPathFromRequest(r)
	store, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	defer store.Close()

	row, err := store.CreateToolset(ctx, req.Name, req.ToolNames)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, ToolsetDTO{
		Name:      row.ID,
		ToolNames: row.ToolNames,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}

func updateToolsetHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/toolset/update")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("toolset_name", name))

	var req UpdateToolsetRequest
	if err := decodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, fmt.Errorf("%w: invalid JSON: %v", errBadRequest, err))
		return
	}

	dbPath := dbPathFromRequest(r)
	store, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	defer store.Close()

	row, err := store.UpdateToolset(ctx, name, req.ToolNames)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	render.JSON(w, r, ToolsetDTO{
		Name:      row.ID,
		ToolNames: row.ToolNames,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}

func deleteToolsetHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/toolset/delete")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("toolset_name", name))

	dbPath := dbPathFromRequest(r)
	store, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	defer store.Close()

	if err := store.DeleteToolset(ctx, name); err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
