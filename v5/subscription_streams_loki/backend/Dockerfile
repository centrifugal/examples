# Use an official Go runtime as a parent image
FROM golang:1.22 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace.
ADD . /app

# Build your program for Linux.
RUN CGO_ENABLED=0 GOOS=linux go build -o backend .

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage to the production image
COPY --from=builder /app/backend .

# Run the myapp binary.
CMD ["./backend"]
