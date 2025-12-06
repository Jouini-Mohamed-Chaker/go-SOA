# Builder
FROM golang:1.25-alpine AS builder

WORKDIR /app

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# copy source and build a static binary
COPY . .

# disable CGO and produce a small static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o loan_service .

# Runner
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/loan_service .

# Copy WSDL file 
COPY loan.wsdl ./loan.wsdl

EXPOSE 8083

CMD ["./loan_service"]