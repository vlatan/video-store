package text

import (
	"github.com/vlatan/video-store/internal/config"
)

type textFiles map[string][]byte

// loadTextFiles parses loads text files into map
func loadTextFiles(cfg *config.Config) textFiles {

	tf := make(textFiles)

	if robotsTxt := buildRobotsTxt(cfg); robotsTxt != nil {
		tf["/robots.txt"] = robotsTxt
	}

	if adsTxt := builAdsTxt(cfg); adsTxt != nil {
		tf["/ads.txt"] = adsTxt
	}

	return tf
}
