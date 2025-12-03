# Builder
FROM golang:1.25-alpine AS builder

WORKDIR /app

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# copy source and build a static binary
COPY . .

# disable CGO and produce a small static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o book_service .

# Runner
FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/book_service .

EXPOSE 8081

CMD ["./book_service"]
