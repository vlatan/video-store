package main

import "factual-docs/internal/worker"

func main() {
	worker := worker.New()
	worker.Run()
}
