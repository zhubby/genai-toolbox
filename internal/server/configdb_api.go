package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// configDBRouter 提供基于 SQLite(ent) 的配置数据 CRUD 接口。
//
// 约定：
// - 默认使用 ~/.toolbox/config.db（storage.Open 的默认逻辑）
// - 可通过 query 参数 dbPath 覆盖数据库路径（用于测试/多实例）
func configDBRouter(s *Server) chi.Router {
	r := chi.NewRouter()

	// Sources：通常用于数据库连接/数据源配置（“数据库配置”）
	r.Route("/sources", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBListSourcesHandler(s, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { configDBCreateSourceHandler(s, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBGetSourceHandler(s, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { configDBUpdateSourceHandler(s, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { configDBDeleteSourceHandler(s, w, r) })
		})
	})

	// Tools
	r.Route("/tools", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBListToolsHandler(s, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { configDBCreateToolHandler(s, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBGetToolHandler(s, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { configDBUpdateToolHandler(s, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { configDBDeleteToolHandler(s, w, r) })
		})
	})

	// AuthServices
	r.Route("/authservices", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBListAuthServicesHandler(s, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { configDBCreateAuthServiceHandler(s, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBGetAuthServiceHandler(s, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { configDBUpdateAuthServiceHandler(s, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { configDBDeleteAuthServiceHandler(s, w, r) })
		})
	})

	// Toolsets
	r.Route("/toolsets", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBListToolsetsHandler(s, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { configDBCreateToolsetHandler(s, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBGetToolsetHandler(s, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { configDBUpdateToolsetHandler(s, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { configDBDeleteToolsetHandler(s, w, r) })
		})
	})

	// Prompts
	r.Route("/prompts", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBListPromptsHandler(s, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { configDBCreatePromptHandler(s, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBGetPromptHandler(s, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { configDBUpdatePromptHandler(s, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { configDBDeletePromptHandler(s, w, r) })
		})
	})

	// Promptsets
	r.Route("/promptsets", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBListPromptsetsHandler(s, w, r) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { configDBCreatePromptsetHandler(s, w, r) })
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { configDBGetPromptsetHandler(s, w, r) })
			r.Put("/", func(w http.ResponseWriter, r *http.Request) { configDBUpdatePromptsetHandler(s, w, r) })
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) { configDBDeletePromptsetHandler(s, w, r) })
		})
	})

	return r
}

