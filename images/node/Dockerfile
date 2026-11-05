FROM node:24.22.2-bookworm-slim@sha256:01d566e8d94d8c619591d78652babfccd3c6992a3dffa9ed5ab39e11cf6c70e9

RUN userdel --remove node \
    && groupadd --gid 1001 node \
    && useradd --uid 1001 --gid node --create-home --shell /usr/sbin/nologin node

LABEL org.opencontainers.image.description="Hardened Debian-slim Node.js base image for building and running Node.js and TypeScript services"

USER node
