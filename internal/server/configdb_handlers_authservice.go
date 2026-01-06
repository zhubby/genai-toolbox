package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func configDBListAuthServicesHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/list")
	defer span.End()

	dbPath := configDBPathFromRequest(r)
	store, ok, err := configDBOpenForRead(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	if !ok {
		render.JSON(w, r, []configDBAuthServiceDTO{})
		return
	}
	defer store.Close()

	rows, err := store.ListAuthServices(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	out := make([]configDBAuthServiceDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, configDBAuthServiceDTO{
			Name:      row.ID,
			Kind:      row.Kind,
			Config:    row.Config,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	render.JSON(w, r, out)
}

func configDBGetAuthServiceHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/get")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("authservice_name", name))

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

	row, err := store.GetAuthService(ctx, name)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	render.JSON(w, r, configDBAuthServiceDTO{
		Name:      row.ID,
		Kind:      row.Kind,
		Config:    row.Config,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}

func configDBCreateAuthServiceHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/create")
	defer span.End()

	var req configDBCreateAuthServiceRequest
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

	row, err := store.CreateAuthService(ctx, req.Name, req.Kind, cfg)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, configDBAuthServiceDTO{
		Name:      row.ID,
		Kind:      row.Kind,
		Config:    row.Config,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}

func configDBUpdateAuthServiceHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/update")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("authservice_name", name))

	var req configDBUpdateAuthServiceRequest
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

	row, err := store.UpdateAuthService(ctx, name, req.Kind, cfg)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	render.JSON(w, r, configDBAuthServiceDTO{
		Name:      row.ID,
		Kind:      row.Kind,
		Config:    row.Config,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}

func configDBDeleteAuthServiceHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.instrumentation.Tracer.Start(r.Context(), "toolbox/server/configdb/authservice/delete")
	defer span.End()

	name := chi.URLParam(r, "name")
	span.SetAttributes(attribute.String("authservice_name", name))

	dbPath := configDBPathFromRequest(r)
	store, err := configDBOpenForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	defer store.Close()

	if err := store.DeleteAuthService(ctx, name); err != nil {
		span.SetStatus(codes.Error, err.Error())
		configDBWriteError(s, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
