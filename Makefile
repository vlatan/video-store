# Simple Makefile for a Go project

build:
	@echo "Building..."	
	@go build -o ./bin/main ./cmd/app
	@go build -o ./bin/worker ./cmd/worker

# Run the application
run:
	@go run ./cmd/app

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f ./bin/main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
		air; \
		echo "Watching...";\
	else \
		read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/air-verse/air@latest; \
			air; \
			echo "Watching...";\
		else \
			echo "You chose not to install air. Exiting..."; \
			exit 1; \
		fi; \
	fi

.PHONY: build run clean watch
