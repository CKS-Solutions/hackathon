# Build context is repo root packages/ (./packages). Copy ms-video app and build.
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY ms-video/app/go.mod ms-video/app/go.sum ./
RUN go mod download

COPY ms-video/app/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/main.go

FROM alpine:3.19

RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 8080
CMD ["./app"]
