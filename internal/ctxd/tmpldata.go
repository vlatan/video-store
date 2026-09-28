package ctxd

import (
	"context"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/ctxv"
)

func GetTmplData(ctx context.Context) *types.TemplateData {
	return ctxv.Get[*types.TemplateData](ctx)
}

func WithTmplData(ctx context.Context, data *types.TemplateData) context.Context {
	return ctxv.WithValue(ctx, data)
}
