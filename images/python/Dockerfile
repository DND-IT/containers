FROM python:3.14.7-slim-bookworm@sha256:416f0db2a2b561945630cef9877a7ea0581b27449eb9fd9df42f03e1b74b5b63

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 1001 python \
    && useradd --uid 1001 --gid python --create-home --shell /usr/sbin/nologin python

LABEL org.opencontainers.image.description="Hardened Debian-slim Python base image for building and running Python services"

USER python
