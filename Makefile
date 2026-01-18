# Define build flags for a static Linux amd64 binary
BUILD_FLAGS = CGO_ENABLED=0 GOOS=linux GOARCH=amd64

build:
	@echo "Building..."
	@${BUILD_FLAGS} go build -o ./bin/app ./cmd/app

# Run the application
run:
	@go run ./cmd/app

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f ./bin/*

.PHONY: build run clean
