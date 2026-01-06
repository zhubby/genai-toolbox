package configdb

import (
	"github.com/googleapis/genai-toolbox/internal/log"
	"go.opentelemetry.io/otel/trace"
)

// Dependencies contains the minimal dependencies needed by the configdb sub-router.
// Keep this small to avoid import cycles with the parent server package.
type Dependencies struct {
	Logger log.Logger
	Tracer trace.Tracer
	// DBPath is the default SQLite config DB path used by configdb handlers.
	DBPath string
}
