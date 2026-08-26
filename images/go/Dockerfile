FROM golang:1.27.0-bookworm@sha256:ded31c68586d2e49e760acc2e65a884b23d032e9bbbed0ae0c55abd3fcaf4452

RUN groupadd --gid 1001 go \
    && useradd --uid 1001 --gid go --create-home --shell /usr/sbin/nologin go \
    && chown -R go:go /go

LABEL org.opencontainers.image.description="Hardened Debian Go toolchain base image for building Go services"

# Build-stage base only: Go compiles to a static binary, so downstream
# Dockerfiles should still multi-stage the compiled binary into their own
# minimal runtime layer rather than shipping this full toolchain to production.
USER go
