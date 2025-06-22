FROM golang:1.23

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies (termasuk driver PostgreSQL)
RUN go mod download

# Copy seluruh source code
COPY . .


# Build binary
RUN go build -o qiscus cmd/main.go

# Expose port jika ada
EXPOSE 8080

# Jalankan aplikasi
CMD ["./qiscus"]
