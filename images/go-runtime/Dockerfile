FROM debian:bookworm-slim@sha256:99c0b25c46d0214d5a2091092e7908907dcaecd6a962cde0ac700acee222b6f3

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 1001 go \
    && useradd --uid 1001 --gid go --create-home --shell /usr/sbin/nologin go

LABEL org.opencontainers.image.description="Minimal Debian-slim runtime base for statically-compiled Go binaries, no toolchain"

USER go
