# Start from the official Golang base image
FROM golang:1.22-alpine as builder

# Install reflex for hot reloading
RUN apk add --no-cache git && go install github.com/cespare/reflex@latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Command to run reflex
CMD reflex -r '\.go$$' -s -- sh -c 'go run ./cmd/api'

# Build the Go app with explicit environment variables
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/api ./cmd/api

# Start a new stage from scratch
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bin/api /root/api

# Also copy invoice.tmpl to the final image
COPY --from=builder /app/invoice.tmpl /root/invoice.tmpl

# Command to run the executable
CMD ["./api"]
