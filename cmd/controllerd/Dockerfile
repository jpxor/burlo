# Use the official Go image as the base
FROM golang:1.19

# Set the working directory inside the container
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY * ./
COPY ../config ./config

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /thermostatd

# Run the application
CMD ["/thermostatd", "-c", "./config/services.toml"]
