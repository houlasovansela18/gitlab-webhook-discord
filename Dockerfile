# Stage 1: Build the Go binary
FROM golang:1.18-alpine AS build

# Set the working directory
WORKDIR /app

# Copy go.mod, then download dependencies
COPY go.mod ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go binary with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go-webhook-service

# Stage 2: Use a minimal base image
FROM alpine:latest

# Install certificates
RUN apk --no-cache add ca-certificates

# Set working directory inside the container
WORKDIR /root/

# Copy the binary from the build stage
COPY --from=build /go-webhook-service /go-webhook-service

# Expose port 3000
EXPOSE 4455

# Run the binary
CMD ["/go-webhook-service"]

