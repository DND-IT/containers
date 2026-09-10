FROM node:24-alpine@sha256:50c8e8ca1d27439048670df5883f32d57cf81cff6233222c893fd0d9884cbd81

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

COPY --chmod=755 entrypoint.sh /usr/local/bin/entrypoint.sh

RUN addgroup -g 1001 argocd \
    && adduser -D -u 1001 -G argocd argocd

USER argocd

EXPOSE 3000

LABEL org.opencontainers.image.description="Argo CD MCP server (applications, sync, and resources over MCP)"

# No CMD: the entrypoint defaults to the HTTP Stream transport on :3000 with a
# GET /healthz probe, taking MCP_PORT, MCP_STATELESS and MCP_ALLOW_UNAUTHENTICATED
# from the environment. Any command passed to the container is exec'd verbatim,
# so `argocd-mcp stdio` and explicit flags keep working.
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
