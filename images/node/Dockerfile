FROM node:24.20.0-bookworm-slim@sha256:ba849c60be29959425b8734d57b8b4b7d56f98edd9504c9af091d5281095a71e

RUN userdel --remove node \
    && groupadd --gid 1001 node \
    && useradd --uid 1001 --gid node --create-home --shell /usr/sbin/nologin node

LABEL org.opencontainers.image.description="Hardened Debian-slim Node.js base image for building and running Node.js and TypeScript services"

USER node
