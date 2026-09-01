FROM python:3.14.7-slim-bookworm@sha256:9ab8d9c8514b44f90cf0029dd42fdd7e9e211e639c8b995304cc04568dee900f

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 1001 python \
    && useradd --uid 1001 --gid python --create-home --shell /usr/sbin/nologin python

LABEL org.opencontainers.image.description="Hardened Debian-slim Python base image for building and running Python services"

USER python
