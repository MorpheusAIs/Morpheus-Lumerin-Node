# Stage 1: Build
# Go 1.26.10. Keep it equal to the toolchain line in go.mod, which CI uses for the release binaries.
FROM golang:1.26-bookworm@sha256:632dff7351799a5a9954a48dc06670ade72ee44e8c351cdc0e5a57af2fea21af AS builder

# Refresh the CA bundle that is copied into the scratch image below.
# - apt-get update fails on any index error (retried for transient mirror errors)
# - ca-certificates is upgraded to the current Debian security release
# - the emSign roots are disabled in /etc/ca-certificates.conf and the bundle is regenerated
# - the regenerated bundle is checked by certificate subject
RUN set -e; \
    ok=; for i in 1 2 3 4 5; do \
      if apt-get update -o APT::Update::Error-Mode=any; then ok=1; break; fi; \
      rm -rf /var/lib/apt/lists/*; sleep 10; \
    done; \
    [ -n "$ok" ]; \
    apt-get install -y --no-install-recommends --only-upgrade ca-certificates; \
    rm -rf /var/lib/apt/lists/*; \
    dpkg --compare-versions "$(dpkg-query -W -f='${Version}' ca-certificates)" ge 20250419~deb12u2; \
    sed -i -E 's|^(mozilla/emSign_[^/]*\.crt)$|!\1|' /etc/ca-certificates.conf; \
    for c in mozilla/emSign_ECC_Root_CA_-_G3.crt mozilla/Comodo_AAA_Services_root.crt; do \
      grep -qxF "!$c" /etc/ca-certificates.conf || echo "!$c" >> /etc/ca-certificates.conf; \
    done; \
    update-ca-certificates --fresh; \
    bundle=/etc/ssl/certs/ca-certificates.crt; \
    openssl crl2pkcs7 -nocrl -certfile "$bundle" | openssl pkcs7 -print_certs -noout | grep '^subject=' > /tmp/ca-subjects.txt; \
    total=$(wc -l < /tmp/ca-subjects.txt); pem=$(grep -c 'BEGIN CERTIFICATE' "$bundle"); \
    echo "CA bundle: $total certificates"; \
    [ "$total" -eq "$pem" ]; \
    [ "$total" -ge 100 ]; \
    if grep -iE 'emSign|AAA Certificate Services' /tmp/ca-subjects.txt; then \
      echo "CA bundle still contains disabled roots" >&2; exit 1; \
    fi; \
    rm -f /tmp/ca-subjects.txt

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