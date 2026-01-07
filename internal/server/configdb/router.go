package configdb

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/codes"
)

// Router 提供基于 SQLite(ent) 的配置数据 CRUD 接口。
//
// 约定：
// - 默认使用 ~/.toolbox/config.db（storage.Open 的默认逻辑）
// - 可通过 query 参数 dbPath 覆盖数据库路径（用于测试/多实例）
func Router(deps Dependencies) chi.Router {
	r := chi.NewRouter()

	// 校验配置库是否可用（可打开/可写）
	r.Post("/validate", func(w http.ResponseWriter, r *http.Request) { validateDBHandler(deps, w, r) })

	// Sources：通常用于数据库连接/数据源配置（“数据库配置”）
	r.Route("/sources", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { listSourcesHandler(deps, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { createSourceHandler(deps, w, r) })
		// 校验 Source 配置是否可用（不落库）
		r.Post("/validate", func(w http.ResponseWriter, r *http.Request) { validateSourceHandler(deps, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { getSourceHandler(deps, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { updateSourceHandler(deps, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { deleteSourceHandler(deps, w, r) })
		})
	})

	// Tools
	r.Route("/tools", func(r chi.Router) {
		r.Get("/kinds", func(w http.ResponseWriter, r *http.Request) { listToolKindsHandler(deps, w, r) })
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { listToolsHandler(deps, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { createToolHandler(deps, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { getToolHandler(deps, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { updateToolHandler(deps, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { deleteToolHandler(deps, w, r) })
		})
	})

	// AuthServices
	r.Route("/authservices", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { listAuthServicesHandler(deps, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { createAuthServiceHandler(deps, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { getAuthServiceHandler(deps, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { updateAuthServiceHandler(deps, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { deleteAuthServiceHandler(deps, w, r) })
		})
	})

	// Toolsets
	r.Route("/toolsets", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { listToolsetsHandler(deps, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { createToolsetHandler(deps, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { getToolsetHandler(deps, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { updateToolsetHandler(deps, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { deleteToolsetHandler(deps, w, r) })
		})
	})

	// Prompts
	r.Route("/prompts", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { listPromptsHandler(deps, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { createPromptHandler(deps, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { getPromptHandler(deps, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { updatePromptHandler(deps, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { deletePromptHandler(deps, w, r) })
		})
	})

	// Promptsets
	r.Route("/promptsets", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { listPromptsetsHandler(deps, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { createPromptsetHandler(deps, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { getPromptsetHandler(deps, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { updatePromptsetHandler(deps, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { deletePromptsetHandler(deps, w, r) })
		})
	})

	return r
}

type ValidateDBResponse struct {
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

func validateDBHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/validate")
	defer span.End()

	dbPath := deps.DBPath
	_, err := openForWrite(ctx, dbPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		// 这里不直接返回 4xx/5xx；按约定返回可用性布尔值与错误信息。
		render.JSON(w, r, ValidateDBResponse{Available: false, Error: fmt.Sprintf("%v", err)})
		return
	}
	render.JSON(w, r, ValidateDBResponse{Available: true})
}
