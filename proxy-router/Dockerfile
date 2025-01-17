# Stage 1: Build
FROM golang:1.23-bullseye AS builder

# Capture the Git tag, commit hash, and architecture
ARG TAG_NAME
ARG COMMIT
ARG TARGETOS
ARG TARGETARCH
ENV TAG_NAME=$TAG_NAME
ENV COMMIT=$COMMIT

WORKDIR /app
COPY . .

# Install dbus
RUN apt-get update && apt-get install -y dbus-x11 && apt-get clean

# Build the Go binary for the target platform
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    TAG_NAME=$TAG_NAME COMMIT=$COMMIT ./build.sh && \
    cp /bin/sh /app/sh && chmod +x /app/sh

# Stage 2: Final Image
FROM scratch
WORKDIR /app

# Copy required files and binary
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/proxy-router /usr/bin/
COPY --from=builder /usr/bin/dbus-launch /usr/bin/

SHELL ["/bin/sh", "-c"]
EXPOSE 3333 8082

ENTRYPOINT ["proxy-router"]