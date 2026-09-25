FROM golang:1.27.1-bookworm@sha256:69a7b9788769bec032d238959b61854e9ae87f57be9029ec04e9885fabf99195

RUN groupadd --gid 1001 go \
    && useradd --uid 1001 --gid go --create-home --shell /usr/sbin/nologin go \
    && chown -R go:go /go

LABEL org.opencontainers.image.description="Hardened Debian Go toolchain base image for building Go services"

# Build-stage base only: Go compiles to a static binary, so downstream
# Dockerfiles should still multi-stage the compiled binary into
# ghcr.io/dnd-it/go-runtime rather than shipping this full toolchain to production.
USER go
