package text

import (
	"fmt"
	"strings"

	"github.com/vlatan/video-store/internal/config"
)

var weirdBots = []string{
	"Nuclei",
	"WikiDo",
	"Riddler",
	"PetalBot",
	"Zoominfobot",
	"Go-http-client",
	"simplecrawler",
	"CazoodleBot",
	"dotbot",
	"Gigabot",
	"Barkrowler",
	"BLEXBot",
	"magpie-crawler",
	"Thinkbot",
}

// buildRobotsTxt build the content of the robots.txt file
func builAdsTxt(cfg *config.Config) []byte {

	if cfg.AdSenseAccount == "" {
		return nil
	}

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
	sitemapURL := fmt.Sprintf("%s://%s/sitemap.xml", cfg.Protocol, cfg.Domain)

	builder.WriteString("# Sitemap\n")
	fmt.Fprintf(&builder, "Sitemap: %s\n\n", sitemapURL)

	builder.WriteString("# Disallow weird bots\n")
	for _, bot := range weirdBots {
		fmt.Fprintf(&builder, "User-agent: %s\n", bot)
	}
	builder.WriteString("Disallow: /\n\n")

	builder.WriteString("# Disallow paths\n")
	builder.WriteString("User-agent: *\n")
	builder.WriteString("Disallow: /auth/\n")
	builder.WriteString("Disallow: /api/")

	return []byte(builder.String())
}
