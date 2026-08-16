# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o routeforge-api ./cmd/api

# Stage 2: Runtime Ultraleve
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/routeforge-api /app/routeforge-api

EXPOSE 8080

CMD ["/app/routeforge-api"]
