# Builder
FROM golang:1.25-alpine AS builder

WORKDIR /app

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# copy source and build a static binary
COPY . .

# disable CGO and produce a small static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user_service .

# Runner
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/user_service .

EXPOSE 8082

CMD ["./user_service"]
