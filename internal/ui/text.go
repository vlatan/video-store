package ui

import (
	"fmt"
	"strings"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/models"
)

var weirdBots = []string{
	"Nuclei",
	"WikiDo",
	"Riddler",
	"PetalBot",
	"Zoominfobot",
	"Go-http-client",
	"Node/simplecrawler",
	"CazoodleBot",
	"dotbot/1.0",
	"Gigabot",
	"Barkrowler",
	"BLEXBot",
	"magpie-crawler",
	"Thinkbot",
}

// GetStaticFiles gets the map containing the static files
func (s *service) TextFiles() models.TextFiles {
	return s.textFiles
}

// PaparseTextFiles parses the text files that app needs and serves
func parseTextFiles(cfg *config.Config) models.TextFiles {
	tf := make(models.TextFiles)
	tf["/robots.txt"] = &models.FileInfo{Bytes: buildRobotsTxt(cfg)}
	tf["/ads.txt"] = &models.FileInfo{Bytes: builAdsTxt(cfg)}
	return tf
}

// buildRobotsTxt build the content of the robots.txt file
func builAdsTxt(cfg *config.Config) []byte {
	return fmt.Appendf(
		[]byte{},
		"google.com, pub-%s, DIRECT, f08c47fec0942fa0",
		cfg.AdSenseAccount,
	)
}

// buildRobotsTxt build the content of the robots.txt file
func buildRobotsTxt(cfg *config.Config) []byte {
	var builder strings.Builder

	// Use canonical domain from config
	sitemapURL := fmt.Sprintf("%s://%s/sitemap.xml",
		cfg.Protocol,
		cfg.Domain,
	)

	builder.WriteString("# Sitemap\n")
	fmt.Fprintf(&builder, "Sitemap: %s\n\n", sitemapURL)

	builder.WriteString("# Ban weird bots\n")
	for _, bot := range weirdBots {
		fmt.Fprintf(&builder, "User-agent: %s\n", bot)
	}
	builder.WriteString("Disallow: /\n\n")

	builder.WriteString("# Disallow all bots on /auth\n")
	builder.WriteString("User-agent: *\n")
	builder.WriteString("Disallow: /auth/")

	return []byte(builder.String())
}
