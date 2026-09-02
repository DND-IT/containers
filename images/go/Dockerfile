FROM golang:1.27.1-bookworm@sha256:648f440f42a0958804efb24df176f806f9d353b41f1c0627f666428e40310f6b

RUN groupadd --gid 1001 go \
    && useradd --uid 1001 --gid go --create-home --shell /usr/sbin/nologin go \
    && chown -R go:go /go

LABEL org.opencontainers.image.description="Hardened Debian Go toolchain base image for building Go services"

# Build-stage base only: Go compiles to a static binary, so downstream
# Dockerfiles should still multi-stage the compiled binary into
# ghcr.io/dnd-it/go-runtime rather than shipping this full toolchain to production.
USER go
