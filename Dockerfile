# Build argument to determine which service to build
# https://docs.docker.com/build/building/variables/#scoping
ARG TARGET="app"

FROM golang:1.25.1-alpine AS builder

# Consume the TARGET build argument in the build stage
ARG TARGET

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

# Consume the TARGET build argument in the build stage
ARG TARGET

# The app will need curl in order to perform the healthcheck
RUN if [ "$TARGET" = "app" ]; then \
    apk add --no-cache curl; \
    fi

# The backup will need the postgresql client in order to dump the DB
RUN if [ "$TARGET" = "backup" ]; then \
    apk add --no-cache postgresql16-client; \
    fi

# Copy the binary from the build stage
COPY --from=builder /src/binary /binary

# Run
CMD ["/binary"]