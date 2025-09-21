ARG YT_DLP_VERSION=2025.08.27-r0
ARG APP_MODULE=github.com/ofstudio/voxify
ARG APP_VERSION=latest
ARG DATA_DIR=/data

# Build stage
FROM golang:1.25-alpine AS builder
ARG APP_MODULE
ARG APP_VERSION

# Copy source code and download dependencies
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Run tests
RUN go test ./...

# Build the application
RUN go build -trimpath \
      -ldflags "-s -w -X ${APP_MODULE}/internal/config.version=${APP_VERSION}" \
      -o /build/voxify-bot ./cmd/voxify-bot

# Final stage
FROM alpine:3.22
ARG YT_DLP_VERSION
ARG DATA_DIR

ENV DATA_DIR=${DATA_DIR}
VOLUME ["${DATA_DIR}"]

ENV DOWNLOAD_DIR=${DATA_DIR}/downloads
ENV PUBLIC_DIR=${DATA_DIR}/public
ENV DB_DIR=${DATA_DIR}/db
ENV DB_FILEPATH=${DB_DIR}/voxify-bot.db

# Install yt-dlp from Alpine repo
RUN apk add --no-cache su-exec yt-dlp=${YT_DLP_VERSION}

# Add entrypoint script and make it executable
COPY entrypoint /
RUN chmod +x /entrypoint

# Copy the binary from the builder stage
COPY --from=builder /build/voxify-bot /

# Start the bot using the entrypoint
ENTRYPOINT ["/entrypoint"]
CMD ["/voxify-bot"]
