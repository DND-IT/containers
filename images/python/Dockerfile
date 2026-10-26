FROM python:3.14.7-slim-bookworm@sha256:82bc3c539b8813ada9d68c63b40158fa002f7f33de9bf3312a3dfdc0620dff56

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 1001 python \
    && useradd --uid 1001 --gid python --create-home --shell /usr/sbin/nologin python

LABEL org.opencontainers.image.description="Hardened Debian-slim Python base image for building and running Python services"

USER python
