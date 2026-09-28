package ctxd

import (
	"context"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/ctxv"
)

// GetTmplData gets template data from context
func GetTmplData(ctx context.Context) *types.TemplateData {
	return ctxv.Get[*types.TemplateData](ctx)
}

// WithTmplData adds template data from context and returns the new context
func WithTmplData(ctx context.Context, data *types.TemplateData) context.Context {
	return ctxv.WithValue(ctx, data)
}
