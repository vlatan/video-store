build:
	@echo "Building..."	
	@go build -o ./bin/app ./cmd/app
	@go build -o ./bin/worker ./cmd/worker
	@go build -o ./bin/backup ./cmd/backup

# Run the application
run:
	@go run ./cmd/app

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f ./bin/*

.PHONY: build run clean
