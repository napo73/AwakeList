# Start from official Go image
FROM golang:1.24.3

WORKDIR /app

# Copy all files
COPY . .

# Download dependencies
RUN go mod tidy

# Build binary (optional)
RUN go build -o auth-service .

# Run directly with go run if you prefer dynamic
CMD ["go", "run", "."]
