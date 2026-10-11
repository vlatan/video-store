package text

import (
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/ui"
)

type textFiles map[string]*types.FileInfo

type Service struct {
	textFiles textFiles
	ui        *ui.Service
}
