package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/backup"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/integrations/cloudflare"
)

func main() {

	// Print separator at the end
	defer fmt.Println(strings.Repeat("-", 70))

	// Give the backup a reasonable time to finish
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		log.Panic(err)
	}

	r2, err := cloudflare.New(ctx, cfg)
	if err != nil {
		log.Panic(err)
	}

	backup := backup.New(cfg, r2)
	if err := backup.Run(ctx); err != nil {
		log.Println(err)
	}
}
