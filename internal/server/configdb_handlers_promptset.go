package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func configDBListPromptsetsHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/list")
	defer span.End()

	dbPath := configDBPathFromRequest(r)
	store, ok, err := configDBOpenForRead(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	if !ok {
		render.JSON(w, r, []configDBPromptsetDTO{})
		return
	}
	defer store.Close()

	rows, err := store.ListPromptsets(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	out := make([]configDBPromptsetDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, configDBPromptsetDTO{
			Name:        row.ID,
			PromptNames: row.PromptNames,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	render.JSON(w, r, out)
}

func configDBGetPromptsetHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/get")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("promptset_name", name))

	dbPath := configDBPathFromRequest(r)
	store, ok, err := configDBOpenForRead(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	if !ok {
		err := fmt.Errorf("config db not found")
		span.SetStatus(codes.Error, err.Error())
		_ = render.Render(w, r, newErrResponse(err, http.StatusNotFound))
		return
	}
	defer store.Close()

	row, err := store.GetPromptset(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	render.JSON(w, r, configDBPromptsetDTO{
		Name:        row.ID,
		PromptNames: row.PromptNames,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
}

func configDBCreatePromptsetHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/create")
	defer span.End()

	var req configDBCreatePromptsetRequest
	if err := configDBDecodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, fmt.Errorf("%w: invalid JSON: %v", errConfigDBBadRequest, err))
		return
	}
	if req.Name == "" {
		err := fmt.Errorf("%w: name is required", errConfigDBBadRequest)
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}

	dbPath := configDBPathFromRequest(r)
	store, err := configDBOpenForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	defer store.Close()

	row, err := store.CreatePromptset(ctx, req.Name, req.PromptNames)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, configDBPromptsetDTO{
		Name:        row.ID,
		PromptNames: row.PromptNames,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
}

func configDBUpdatePromptsetHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/update")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("promptset_name", name))

	var req configDBUpdatePromptsetRequest
	if err := configDBDecodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, fmt.Errorf("%w: invalid JSON: %v", errConfigDBBadRequest, err))
		return
	}

	dbPath := configDBPathFromRequest(r)
	store, err := configDBOpenForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	defer store.Close()

	row, err := store.UpdatePromptset(ctx, name, req.PromptNames)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	render.JSON(w, r, configDBPromptsetDTO{
		Name:        row.ID,
		PromptNames: row.PromptNames,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
}

func configDBDeletePromptsetHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/promptset/delete")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("promptset_name", name))

	dbPath := configDBPathFromRequest(r)
	store, err := configDBOpenForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	defer store.Close()

	if err := store.DeletePromptset(ctx, name); err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
