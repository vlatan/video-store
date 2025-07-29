FROM golang:1.24-alpine AS builder

# Build argument to determine which service to build
ARG TARGET="app"

# Set destination for COPY
WORKDIR /src

# Copy the source code
COPY go.mod go.sum ./
COPY web ./web
COPY internal ./internal
COPY cmd ./cmd

# Download Go modules
RUN go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o binary ./cmd/${TARGET}

# Use small image for the final stage
FROM alpine:3.21

# Redeclare the ARG for this stage
ARG TARGET="app"

# The backup app will need the postgresql client in order to dump the DB
RUN if [ "$TARGET" = "backup" ]; then \
    apk update && apk add postgresql16-client; \
    fi

# Copy the binary from the build stage
COPY --from=builder /src/binary /binary

# Run
CMD ["/binary"]