package text

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/types"
)

// loadTextFiles parses loads text files into map
func loadTextFiles(cfg *config.Config) map[string]*types.FileInfo {

	tf := make(map[string]*types.FileInfo)

	if robotsTxt := buildRobotsTxt(cfg); robotsTxt != nil {
		tf["/robots.txt"] = &types.FileInfo{Bytes: robotsTxt}
	}

	if adsTxt := builAdsTxt(cfg); adsTxt != nil {
		tf["/ads.txt"] = &types.FileInfo{Bytes: adsTxt}
	}

	return tf
}
