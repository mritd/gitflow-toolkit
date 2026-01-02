# Test container for gitflow-toolkit installation testing
FROM golang:1.24-alpine

# Install dependencies
RUN apk add --no-cache git bash sudo shadow

# Create test user (non-root)
RUN adduser -D -s /bin/bash testuser && \
    echo "testuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Set up working directory
WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o /tmp/gitflow-toolkit .

# Set up git config for test user
USER testuser
RUN git config --global user.name "Test User" && \
    git config --global user.email "test@example.com" && \
    git config --global init.defaultBranch main

WORKDIR /home/testuser

# Default command runs the test script
CMD ["/bin/bash"]
