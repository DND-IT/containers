FROM ghcr.io/dnd-it/node:0.1.0@sha256:d7caad1db04c9c6bab57e3a2a8f63dfda67f9cc53f9cda93f3fde18ee040c7ed

# renovate: datasource=npm depName=@anthropic-ai/claude-code
ARG CLAUDE_CODE_VERSION=2.1.310
# The node base image ships an older bundled npm whose own vendored deps carry
# known CVEs; upgrading npm itself is the only way to move them.
# renovate: datasource=npm depName=npm
ARG NPM_VERSION=12.2.1
# renovate: datasource=github-releases depName=cli/cli
ARG GH_VERSION=2.103.2
# renovate: datasource=github-releases depName=astral-sh/uv
ARG UV_VERSION=0.13.5
ARG TARGETARCH

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

USER root

# tmux keeps an interactive session alive across a dropped exec connection
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        jq \
        less \
        openssh-client \
        python3 \
        ripgrep \
        tmux \
        unzip \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_${TARGETARCH}.tar.gz" \
        | tar -xz -C /usr/local/bin --strip-components=2 --no-same-owner "gh_${GH_VERSION}_linux_${TARGETARCH}/bin/gh" \
    && UV_ARCH="$(case "${TARGETARCH}" in amd64) echo x86_64 ;; arm64) echo aarch64 ;; esac)" \
    && curl -fsSL "https://github.com/astral-sh/uv/releases/download/${UV_VERSION}/uv-${UV_ARCH}-unknown-linux-gnu.tar.gz" \
        | tar -xz -C /usr/local/bin --strip-components=1 --no-same-owner \
    && gh --version \
    && uv --version

# npm 12 blocks dependency install scripts by default; rather than allowing them
# globally, run the one Claude Code needs, which links its platform-native binary.
RUN npm install -g "npm@${NPM_VERSION}" \
    && npm install -g "@anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}" \
    && node "$(npm root -g)/@anthropic-ai/claude-code/install.cjs" \
    && npm cache clean --force \
    && test "$(claude --version)" = "${CLAUDE_CODE_VERSION} (Claude Code)"

COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint

RUN mkdir /workspace \
    && chown node:node /workspace

USER node

ENV AGENT_WORKSPACE=/workspace \
    CLAUDE_CONFIG_DIR=/home/node/.claude \
    DISABLE_AUTOUPDATER=1

WORKDIR /workspace

LABEL org.opencontainers.image.description="Headless Claude Code agent with git-repo and agent-profile bootstrap"

ENTRYPOINT ["/usr/local/bin/entrypoint"]
CMD ["claude", "-p"]
