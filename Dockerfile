ARG YT_DLP_VERSION=2025.08.27-r0
ARG APP_MODULE=github.com/ofstudio/voxify
ARG APP_VERSION=latest

FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./...
ARG APP_MODULE
ARG APP_VERSION
RUN go build -trimpath \
      -ldflags "-s -w -X ${APP_MODULE}/internal/config.version=${APP_VERSION}" \
      -o /build/voxify-bot ./cmd/voxify-bot

FROM alpine:3.22
ARG YT_DLP_VERSION
RUN apk add --no-cache yt-dlp=${YT_DLP_VERSION}
COPY --from=builder /build/voxify-bot /
EXPOSE 8080
VOLUME ["/data/db", "/data/public", "/data/downloads"]
ENV DB_FILEPATH=/data/db/voxify-bot.db
ENV DOWNLOAD_DIR=/data/downloads
ENV PUBLIC_DIR=/data/public
CMD ["/voxify-bot"]
