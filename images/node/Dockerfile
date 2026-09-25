FROM node:24.21.0-bookworm-slim@sha256:0e0ff40c39bc087845bfb27465a0df4ea419520094bc35842ff83dd8cbe6f9b6

RUN userdel --remove node \
    && groupadd --gid 1001 node \
    && useradd --uid 1001 --gid node --create-home --shell /usr/sbin/nologin node

LABEL org.opencontainers.image.description="Hardened Debian-slim Node.js base image for building and running Node.js and TypeScript services"

USER node
