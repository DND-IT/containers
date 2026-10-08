FROM golang:1.27.2-bookworm@sha256:5cf287a799e6b94384bad13d16b14904c531f51ba65792237e122ce42b392f61

RUN groupadd --gid 1001 go \
    && useradd --uid 1001 --gid go --create-home --shell /usr/sbin/nologin go \
    && chown -R go:go /go

LABEL org.opencontainers.image.description="Hardened Debian Go toolchain base image for building Go services"

# Build-stage base only: Go compiles to a static binary, so downstream
# Dockerfiles should still multi-stage the compiled binary into
# ghcr.io/dnd-it/go-runtime rather than shipping this full toolchain to production.
USER go
