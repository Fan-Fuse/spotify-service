# Use the official Go image as the base image
FROM golang:1.21.4

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download and install the project dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application, make sure CGO_ENABLED=1 is set for cgo
RUN CGO_ENABLED=1 go build -o app

# Set the entry point for the container
ENTRYPOINT ["./app"]