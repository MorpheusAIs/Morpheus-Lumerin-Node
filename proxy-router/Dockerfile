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
RUN if [ "$TARGETOS" = "windows" ]; then \
      go build -o /app/bin/proxy-router.exe ./cmd/main.go; \
    else \
      CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
      go build -o /app/bin/proxy-router ./cmd/main.go; \
    fi

# Use runtime images for Linux and Windows
FROM scratch AS runtime-linux
WORKDIR /app
COPY --from=builder /app/bin/proxy-router /usr/bin/
COPY --from=builder-linux /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["proxy-router"]

FROM mcr.microsoft.com/windows/nanoserver:ltsc2022 AS runtime-windows
WORKDIR /app
COPY --from=builder /app/bin/proxy-router.exe /usr/bin/
ENTRYPOINT ["proxy-router.exe"]

# Use ARG to select the runtime stage
FROM runtime-${TARGETOS} AS runtime