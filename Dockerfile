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

# The app and worker base 
FROM jauderho/yt-dlp AS app
FROM jauderho/yt-dlp AS worker

# The backup will need the postgresql client in order to dump the DB
FROM alpine:3.21 AS backup
RUN apk add --no-cache postgresql16-client

# Final stage - pick the right base
FROM ${TARGET}

COPY --from=builder /src/binary /binary

# Run
CMD ["/binary"]