# Common build arguments
ARG TARGETOS
ARG TARGETARCH

# Use Bullseye for Linux
FROM golang:1.23-bullseye AS builder-linux

# Use Windows Server Core for Windows
FROM golang:1.23-windowsservercore AS builder-windows

# Dynamically choose the builder base image
FROM builder-${TARGETOS} AS builder

ARG TAG_NAME
ARG COMMIT
ENV TAG_NAME=$TAG_NAME
ENV COMMIT=$COMMIT

WORKDIR /app
COPY . .

# Build for the target platform
RUN if [ "$TARGETOS" = "linux" ]; then \
      CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /app/bin/proxy-router ./cmd/main.go; \
    elif [ "$TARGETOS" = "windows" ]; then \
      powershell -Command "go build -o /app/bin/proxy-router.exe ./cmd/main.go"; \
    fi

# Runtime image for Linux
FROM scratch AS runtime-linux
WORKDIR /app
COPY --from=builder /app/bin/proxy-router /usr/bin/
COPY --from=builder-linux /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["proxy-router"]

# Runtime image for Windows
FROM mcr.microsoft.com/windows/nanoserver:ltsc2022 AS runtime-windows
WORKDIR /app
COPY --from=builder /app/bin/proxy-router.exe /usr/bin/
ENTRYPOINT ["proxy-router.exe"]

# Select the runtime image based on TARGETOS
FROM runtime-${TARGETOS} AS runtime