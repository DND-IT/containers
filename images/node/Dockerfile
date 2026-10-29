FROM node:24.22.2-bookworm-slim@sha256:c86ecead9e55a5210ac1727de38a4ca840bf03dcaf34dc9749addbf383b1920c

RUN userdel --remove node \
    && groupadd --gid 1001 node \
    && useradd --uid 1001 --gid node --create-home --shell /usr/sbin/nologin node

LABEL org.opencontainers.image.description="Hardened Debian-slim Node.js base image for building and running Node.js and TypeScript services"

USER node
