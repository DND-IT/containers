FROM golang:1.27.3-bookworm@sha256:d9b6a1beb9474adc2a1ce842573639b44b2a0dd49b6669f51d3821baf9ec37c9

RUN groupadd --gid 1001 go \
    && useradd --uid 1001 --gid go --create-home --shell /usr/sbin/nologin go \
    && chown -R go:go /go

LABEL org.opencontainers.image.description="Hardened Debian Go toolchain base image for building Go services"

# Build-stage base only: Go compiles to a static binary, so downstream
# Dockerfiles should still multi-stage the compiled binary into
# ghcr.io/dnd-it/go-runtime rather than shipping this full toolchain to production.
USER go
