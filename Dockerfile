# Use the official Golang image to build the Go binary
FROM golang:1.23 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Install air for live reloading
RUN go install github.com/air-verse/air@latest

# Start a new stage from scratch
FROM golang:1.23 AS dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests and source code
COPY go.mod go.sum ./
COPY . .

# Install air for live reloading
RUN go install github.com/air-verse/air@latest

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run air (live reloading)
CMD ["air"]
