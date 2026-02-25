# Build context is repo root packages/ (./packages). Build stub and run as ms-auth.
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY ms-notify/app/go.mod ms-notify/app/go.sum ./
RUN go mod download

COPY ms-notify/app/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/main.go

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 8080
CMD ["./app"]
