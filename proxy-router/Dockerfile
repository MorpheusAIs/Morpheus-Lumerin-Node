# Stage 1: Build
FROM golang:1.22.3-alpine as builder

# Capture the Git tag and commit hash during build
ARG TAG_NAME
ARG COMMIT
ENV TAG_NAME=$TAG_NAME
ENV COMMIT=$COMMIT

WORKDIR /app
COPY . .

# Build the Go binary (recommended for linux/amd64...for MacARM use buildx)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  TAG_NAME=$TAG_NAME COMMIT=$COMMIT ./build.sh && \
  cp /bin/sh /app/sh && chmod +x /app/sh

  # Multiplatform Build Notes:   
# to support both amd64 and arm64, use Docker’s Buildx to create a multi-architecture image
# docker buildx create --use
# docker buildx build --platform linux/amd64,linux/arm64 -t proxy-router:latest .

# Stage 2: Final Image
FROM scratch
WORKDIR /app

# Copy required files and binary
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/proxy-router /usr/bin/

# Optional Copy utilities from busybox image
# COPY --from=busybox /bin /bin
# COPY --from=busybox /lib /lib

SHELL ["/bin/sh", "-c"]
EXPOSE 3333 8082

ENTRYPOINT ["proxy-router"]