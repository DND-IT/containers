FROM node:24-alpine@sha256:a0b9bf06e4e6193cf7a0f58816cc935ff8c2a908f81e6f1a95432d679c54fbfd

# renovate: datasource=npm depName=argocd-mcp
ARG ARGOCD_MCP_VERSION=0.9.0
# The node base image ships an older bundled npm whose own vendored deps carry
# known CVEs; upgrading npm itself is the only way to move them.
# renovate: datasource=npm depName=npm
ARG NPM_VERSION=12.0.2

# The published bundle is architecture-independent JavaScript, so no TARGETARCH
# mapping is needed — the only per-arch content comes from the base image.
RUN npm install -g "npm@${NPM_VERSION}" \
    && npm install -g "argocd-mcp@${ARGOCD_MCP_VERSION}" \
    && npm cache clean --force \
    && test "$(argocd-mcp --version)" = "${ARGOCD_MCP_VERSION}"

RUN addgroup -g 1001 argocd \
    && adduser -D -u 1001 -G argocd argocd

USER argocd

EXPOSE 3000

LABEL org.opencontainers.image.description="Argo CD MCP server (applications, sync, and resources over MCP)"

# HTTP Stream transport, listening on :3000 with a GET /healthz probe. Override
# the command with `stdio` for clients that speak MCP over stdin/stdout, and add
# --stateless when running more than one replica without sticky sessions.
CMD ["argocd-mcp", "http"]
