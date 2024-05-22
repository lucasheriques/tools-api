# Start from the official Golang base image
FROM golang:1.22 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY ./go.mod ./go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY .. .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux make build/api

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
