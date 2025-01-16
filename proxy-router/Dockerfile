# Base image selection based on OS
ARG TARGETOS
ARG TARGETARCH

# Use Bullseye for Linux builds
FROM golang:1.23-bullseye AS builder-linux

# Use Windows Server Core for Windows builds
FROM golang:1.23-windowsservercore AS builder-windows

# Select the correct base image dynamically
FROM builder-${TARGETOS} AS builder

# Common build arguments
ARG TAG_NAME
ARG COMMIT
ENV TAG_NAME=$TAG_NAME
ENV COMMIT=$COMMIT

WORKDIR /app
COPY . .

# Build for the target platform
RUN if [ "$TARGETOS" = "windows" ]; then \
      go build -o /app/bin/proxy-router.exe ./cmd/main.go; \
    else \
      CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
      go build -o /app/bin/proxy-router ./cmd/main.go; \
    fi

# Runtime image selection
FROM ${TARGETOS} == "windows" ? mcr.microsoft.com/windows/nanoserver:ltsc2022 : scratch AS runtime

WORKDIR /app

# Copy the built binary and certificates
COPY --from=builder /app/bin/* /usr/bin/
COPY --from=builder-linux /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the entrypoint based on OS
ENTRYPOINT ["proxy-router"]