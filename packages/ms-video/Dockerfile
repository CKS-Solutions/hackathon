# Build context is repo root packages/ (./packages). Build stub and run as ms-video.
FROM golang:1.24-alpine AS build
WORKDIR /app

COPY ms-stub/go.mod ms-stub/go.sum* ./
RUN go mod download 2>/dev/null || true
COPY ms-stub/ .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /ms-stub .

FROM alpine:3.19
RUN apk add --no-cache wget && \
    addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

WORKDIR /app
COPY --from=build --chown=appuser:appgroup /ms-stub .

USER 1001
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q -O- http://localhost:8080/health || exit 1

CMD ["/app/ms-stub"]
