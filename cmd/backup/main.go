package main

import (
	"context"
	"factual-docs/internal/backup"
	"log"
)

func main() {
	ctx := context.Background()
	backupService := backup.New(ctx)
	if err := backupService.Run(ctx); err != nil {
		log.Println(err)
	}
}
