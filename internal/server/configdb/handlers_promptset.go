package configdb

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func listPromptsetsHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/list")
	defer span.End()

	dbPath := deps.DBPath
	store, ok, err := openForRead(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	if !ok {
		render.JSON(w, r, []PromptsetDTO{})
		return
	}

	rows, err := store.ListPromptsets(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	out := make([]PromptsetDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, PromptsetDTO{
			Name:        row.ID,
			PromptNames: row.PromptNames,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	render.JSON(w, r, out)
}

func getPromptsetHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/get")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("promptset_name", name))

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

	row, err := store.GetPromptset(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	render.JSON(w, r, PromptsetDTO{
		Name:        row.ID,
		PromptNames: row.PromptNames,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
}

func createPromptsetHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/create")
	defer span.End()

	var req CreatePromptsetRequest
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

	dbPath := deps.DBPath
	store, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}

	row, err := store.CreatePromptset(ctx, req.Name, req.PromptNames)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, PromptsetDTO{
		Name:        row.ID,
		PromptNames: row.PromptNames,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
}

func updatePromptsetHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/update")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("promptset_name", name))

	var req UpdatePromptsetRequest
	if err := decodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, fmt.Errorf("%w: invalid JSON: %v", errBadRequest, err))
		return
	}

	dbPath := deps.DBPath
	store, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}

	row, err := store.UpdatePromptset(ctx, name, req.PromptNames)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	render.JSON(w, r, PromptsetDTO{
		Name:        row.ID,
		PromptNames: row.PromptNames,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
}

func deletePromptsetHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/delete")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("promptset_name", name))

	dbPath := deps.DBPath
	store, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}

	if err := store.DeletePromptset(ctx, name); err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
