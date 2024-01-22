# Start from the latest Golang base image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
# Adjust the path according to your main package location
RUN go build -o main ./cmd/integral/main.go

# Expose port 8080 (or any other port your app uses)
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
