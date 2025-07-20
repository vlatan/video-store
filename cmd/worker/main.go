package main

import (
	"factual-docs/internal/worker"
	"log"
)

func main() {
	worker := worker.New()
	if err := worker.Run(); err != nil {
		log.Println(err)
	}
}
