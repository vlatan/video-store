package main

import (
	"context"
	"factual-docs/internal/backup"
	"factual-docs/internal/config"
	"factual-docs/internal/r2"
	"log"
	"time"
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
