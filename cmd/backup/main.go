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

	// Give the backup a reasonable time to finish
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := runBackup(ctx); err != nil {
		log.Printf("Backup error: %v", err)
	}

	fmt.Println(strings.Repeat("-", 70))
}

// runBackup runs the backup service
func runBackup(ctx context.Context) error {

	cfg, err := config.New()
	if err != nil {
		return err
	}

	r2, err := cloudflare.New(ctx, cfg)
	if err != nil {
		return err
	}

	backup := backup.New(cfg, r2)
	return backup.Run(ctx)
}
