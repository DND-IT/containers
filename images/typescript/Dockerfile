FROM node:24.19.0-bookworm-slim@sha256:a9f5f7c91a432850b2a8a7797adf5eadb6c733ceed61167806cee7ea7fbc29df

# renovate: datasource=npm depName=typescript
ARG TYPESCRIPT_VERSION=7.0.2

RUN groupadd --gid 1001 typescript \
    && useradd --uid 1001 --gid typescript --create-home --shell /usr/sbin/nologin typescript

RUN npm install -g "typescript@${TYPESCRIPT_VERSION}" \
    && npm cache clean --force \
    && test "$(tsc --version)" = "Version ${TYPESCRIPT_VERSION}"

LABEL org.opencontainers.image.description="Hardened Debian-slim Node.js and TypeScript base image for building and running TypeScript services"

USER typescript
