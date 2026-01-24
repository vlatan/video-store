package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/backup"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/integrations/r2"
)

func main() {

	// Print separator at the end
	defer fmt.Println(strings.Repeat("-", 70))

	// Give the backup a reasonable time to finish
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cfg := config.New()
	r2Service := r2.New(ctx, cfg)
	backupService := backup.New(cfg, r2Service)
	if err := backupService.Run(ctx); err != nil {
		log.Println(err)
	}
}
