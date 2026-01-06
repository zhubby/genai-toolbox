package configdb

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"github.com/googleapis/genai-toolbox/internal/util"
	"go.opentelemetry.io/otel/codes"
)

type ValidateSourceResponse struct {
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

func validateSourceHandler(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	ctx, span := deps.Tracer.Start(r.Context(), "toolbox/server/configdb/source/validate")
	defer span.End()

	var req ValidateSourceRequest
	if err := decodeJSON(r, &req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, fmt.Errorf("%w: invalid JSON: %v", errBadRequest, err))
		return
	}
	if req.Kind == "" {
		err := fmt.Errorf("%w: kind is required", errBadRequest)
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, err)
		return
	}

	cfg, err := normalizeNumbers(req.Config)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		writeError(deps, w, r, fmt.Errorf("%w: %v", errBadRequest, err))
		return
	}

	// sources.DecodeConfig 的解码/校验依赖 config map 中包含 kind/name 等字段。
	// 这里补齐最小字段，便于用户仅传“连接参数”即可验证。
	full := map[string]any{
		"kind": req.Kind,
		"name": "validate",
	}
	for k, v := range cfg {
		full[k] = v
	}

	dec, err := util.NewStrictDecoder(full)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		render.JSON(w, r, ValidateSourceResponse{Available: false, Error: err.Error()})
		return
	}
	sourceCfg, err := sources.DecodeConfig(ctx, req.Kind, "validate", dec)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		render.JSON(w, r, ValidateSourceResponse{Available: false, Error: err.Error()})
		return
	}

	src, err := sourceCfg.Initialize(ctx, deps.Tracer)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		render.JSON(w, r, ValidateSourceResponse{Available: false, Error: err.Error()})
		return
	}

	// Best-effort 关闭资源（部分 source 支持 Close）
	if c, ok := src.(interface{ Close() error }); ok {
		_ = c.Close()
	}

	render.JSON(w, r, ValidateSourceResponse{Available: true})
}
