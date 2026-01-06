package configdb

import (
	"net/http"

	"github.com/go-chi/render"
)

// errResponse mirrors internal/server/api.go's error response shape, but lives in this package
// to avoid import cycles.
type errResponse struct {
	Err            error `json:"-"`
	HTTPStatusCode int   `json:"-"`

	StatusText string `json:"status"`
	ErrorText  string `json:"error,omitempty"`
}

func newErrResponse(err error, code int) *errResponse {
	return &errResponse{
		Err:            err,
		HTTPStatusCode: code,
		StatusText:     http.StatusText(code),
		ErrorText:      err.Error(),
	}
}

func (e *errResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

var _ render.Renderer = (*errResponse)(nil)
