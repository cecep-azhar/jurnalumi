# Multi-stage build for production
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/server ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/web /app/web

ENV APP_ENV=production
ENV PORT=8085

EXPOSE 8085

ENTRYPOINT ["/app/server"]
