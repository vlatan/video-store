package static

import "github.com/vlatan/video-store/internal/types"

type staticFiles map[string]*types.FileInfo

type Service struct {
	staticFiles staticFiles
}
