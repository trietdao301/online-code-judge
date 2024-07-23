# Start from the official Go image
FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download && go mod verify

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -v -o /usr/local/bin/app ./main.go

EXPOSE 8080
CMD ["app"]


