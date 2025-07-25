package main

import (
	"context"
	"factual-docs/internal/worker"
	"log"
)

func main() {
	worker := worker.New()
	ctx := context.Background()
	if err := worker.Run(ctx); err != nil {
		log.Println(err)
	}
}
