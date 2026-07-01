package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/backup"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/utils"
)

func main() {

	// Print separator at the end
	defer utils.LogPlainln(strings.Repeat("-", 70))

	// Give the backup a reasonable time to finish
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	r2Service, err := r2.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	backupService := backup.New(cfg, r2Service)
	if err := backupService.Run(ctx); err != nil {
		log.Println(err)
	}
}
