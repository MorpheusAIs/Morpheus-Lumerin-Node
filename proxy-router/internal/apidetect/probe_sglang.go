package apidetect

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

func (d *Detector) probeSGLang(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	info := d.getJSON(ctx, base+"/get_model_info", apiKey)
	if info == nil {
		return false
	}
	path, ok := info["model_path"].(string)
	if !ok {
		return false
	}
	ev.Stack = "sglang"
	ev.ServedModelID = path
	tracef(ctx, "identified sglang by /get_model_info; served model path %q", path)
	return true
}
