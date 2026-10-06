FROM debian:bookworm-slim@sha256:7c7b2c966bc9ee8cedfeef67e0e279108992c77681fa595db4a9d65c06ccc587

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 1001 go \
    && useradd --uid 1001 --gid go --create-home --shell /usr/sbin/nologin go

LABEL org.opencontainers.image.description="Minimal Debian-slim runtime base for statically-compiled Go binaries, no toolchain"

USER go
