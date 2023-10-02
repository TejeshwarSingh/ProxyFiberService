# Use the official Golang image as a base image
FROM golang:1.19

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go modules manifests and download the dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o main .

# Expose port 3100 to the outside world
EXPOSE 3100

# Command to run the application
CMD ["./main"]
