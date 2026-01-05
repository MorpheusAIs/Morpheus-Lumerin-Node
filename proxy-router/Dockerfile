# Stage 1: Build
FROM golang:1.23-bullseye AS builder

# Create and set permissions for /tmp directory
RUN mkdir -p /tmp && chmod 1777 /tmp

# Capture the Git tag, commit hash, and architecture
ARG TAG_NAME
ARG COMMIT
ARG TARGETOS
ARG TARGETARCH
ENV TAG_NAME=$TAG_NAME
ENV COMMIT=$COMMIT

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code after dependencies are cached
COPY . .

# Ensure go.sum is up to date (handles cases where new deps were added but go.sum wasn't updated)
RUN go mod tidy

# Build the Go binary for the target platform
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    TAG_NAME=$TAG_NAME COMMIT=$COMMIT ./build.sh && \
    cp /bin/sh /app/sh && chmod +x /app/sh

# Stage 2: Final Image
FROM scratch
WORKDIR /app

# Copy required files and binary
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/proxy-router /usr/bin/
# COPY --from=builder /usr/bin/dbus-launch /usr/bin/

# Copy /tmp directory from builder
COPY --from=builder /tmp /tmp

SHELL ["/bin/sh", "-c"]
EXPOSE 3333 8082

ENTRYPOINT ["proxy-router"] 