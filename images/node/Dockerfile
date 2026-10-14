FROM node:24.22.0-bookworm-slim@sha256:a5f4366e834695d5f363b920c104953a535d3a9864d1bdf74566a0ebe01daee5

RUN userdel --remove node \
    && groupadd --gid 1001 node \
    && useradd --uid 1001 --gid node --create-home --shell /usr/sbin/nologin node

LABEL org.opencontainers.image.description="Hardened Debian-slim Node.js base image for building and running Node.js and TypeScript services"

USER node
