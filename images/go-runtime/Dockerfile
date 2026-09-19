FROM debian:bookworm-slim@sha256:3783cc01769c7b2b1b83a5c5ad96c815348e28ed7da68e2e3687004faa906251

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 1001 go \
    && useradd --uid 1001 --gid go --create-home --shell /usr/sbin/nologin go

LABEL org.opencontainers.image.description="Minimal Debian-slim runtime base for statically-compiled Go binaries, no toolchain"

USER go
