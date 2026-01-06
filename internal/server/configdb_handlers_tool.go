package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func configDBListToolsHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/tool/list")
	defer span.End()

	dbPath := configDBPathFromRequest(r)
	store, ok, err := configDBOpenForRead(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	if !ok {
		render.JSON(w, r, []configDBToolDTO{})
		return
	}
	defer store.Close()

	rows, err := store.ListTools(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	out := make([]configDBToolDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, configDBToolDTO{
			Name:       row.ID,
			Kind:       row.Kind,
			SourceName: row.SourceName,
			Config:     row.Config,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		})
	}
	render.JSON(w, r, out)
}

func configDBGetToolHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/tool/get")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("tool_name", name))

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

	row, err := store.GetTool(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	render.JSON(w, r, configDBToolDTO{
		Name:       row.ID,
		Kind:       row.Kind,
		SourceName: row.SourceName,
		Config:     row.Config,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	})
}

func configDBCreateToolHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/tool/create")
	defer span.End()

	var req configDBCreateToolRequest
	if err := configDBDecodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, fmt.Errorf("%w: invalid JSON: %v", errConfigDBBadRequest, err))
		return
	}
	if req.Name == "" || req.Kind == "" {
		err := fmt.Errorf("%w: name/kind is required", errConfigDBBadRequest)
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	cfg, err := configDBNormalizeNumbers(req.Config)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, fmt.Errorf("%w: %v", errConfigDBBadRequest, err))
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

	row, err := store.CreateTool(ctx, req.Name, req.Kind, req.SourceName, cfg)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, configDBToolDTO{
		Name:       row.ID,
		Kind:       row.Kind,
		SourceName: row.SourceName,
		Config:     row.Config,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	})
}

func configDBUpdateToolHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/tool/update")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("tool_name", name))

	var req configDBUpdateToolRequest
	if err := configDBDecodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, fmt.Errorf("%w: invalid JSON: %v", errConfigDBBadRequest, err))
		return
	}
	if req.Kind == "" {
		err := fmt.Errorf("%w: kind is required", errConfigDBBadRequest)
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	cfg, err := configDBNormalizeNumbers(req.Config)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, fmt.Errorf("%w: %v", errConfigDBBadRequest, err))
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

	row, err := store.UpdateTool(ctx, name, req.Kind, req.SourceName, cfg)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	render.JSON(w, r, configDBToolDTO{
		Name:       row.ID,
		Kind:       row.Kind,
		SourceName: row.SourceName,
		Config:     row.Config,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	})
}

func configDBDeleteToolHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/tool/delete")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("tool_name", name))

	dbPath := configDBPathFromRequest(r)
	store, err := configDBOpenForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	defer store.Close()

	if err := store.DeleteTool(ctx, name); err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
