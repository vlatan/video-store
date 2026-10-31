# Define build flags for a static Linux amd64 binary
BUILD_FLAGS = CGO_ENABLED=0 GOOS=linux GOARCH=amd64

.PHONY: build
build:
	@echo "building..."
	@${BUILD_FLAGS} go build -o ./bin/app ./cmd/app

.PHONY: run 
run: build
	@./bin/app

.PHONY: clean 
clean:
	@echo "cleaning..."
	@go clean
	@rm -f ./bin/*

