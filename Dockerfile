# Build argument to determine which service to build
# https://docs.docker.com/build/building/variables/#scoping
ARG TARGET="app"

FROM golang:1.25-alpine AS builder

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
FROM alpine:3.21 AS alpine-base


# The app will need curl in order to perform the healthcheck
FROM alpine-base AS app
RUN apk add --no-cache curl


# The backup will need the postgresql client in order to dump the DB
FROM alpine-base AS backup
RUN apk add --no-cache postgresql16-client


# Worker needs yt-dlp, deno, ffmpeg
FROM debian:bookworm-slim AS worker
RUN set -e; \
    apt-get update && apt-get install -y --no-install-recommends curl ffmpeg python3 ca-certificates unzip; \
    curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp; \
    chmod a+rx /usr/local/bin/yt-dlp; \
    curl -fsSL https://deno.land/install.sh | DENO_INSTALL=/usr/local sh; \
    rm -rf /var/lib/apt/lists/*


# Final stage - pick the right base
FROM ${TARGET}

COPY --from=builder /src/binary /binary

# Run
CMD ["/binary"]