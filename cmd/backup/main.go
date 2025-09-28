package main

import (
	"context"
	"log"
	"time"

	"github.com/vlatan/video-store/internal/backup"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/integrations/r2"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cfg := config.New()
	r2Service := r2.New(ctx, cfg)
	backupService := backup.New(cfg, r2Service)
	if err := backupService.Run(ctx); err != nil {
		log.Println(err)
	}
}
