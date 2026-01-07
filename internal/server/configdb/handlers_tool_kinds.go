package configdb

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"go.opentelemetry.io/otel/codes"
)

func listToolKindsHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	_, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/tool/kinds")
	defer span.End()

	prefix := r.URL.Query().Get("prefix")
	kinds := tools.RegisteredKinds(prefix)
	if kinds == nil {
		kinds = []string{}
	}
	span.SetStatus(codes.Ok, "")
	render.JSON(w, r, kinds)
}


