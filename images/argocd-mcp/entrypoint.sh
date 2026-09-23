#!/bin/sh
# Upstream exposes the port, stateless and allow-unauthenticated settings as
# flags only. Translate them from the environment so the image runs unchanged
# on platforms that cannot override the container command.
set -eu

# On platforms whose environment variables are not secret (AgentCore Runtime),
# credentials come from a Secrets Manager secret the execution role can read.
if [ -n "${SECRET_ENV_ARN:-}" ]; then
    exports=$(node /opt/secret-env/secret-env.mjs)
    eval "$exports"
    unset exports
fi

if [ "$#" -gt 0 ]; then
    exec "$@"
fi

set -- argocd-mcp http

if [ -n "${MCP_PORT:-}" ]; then
    set -- "$@" --port="$MCP_PORT"
fi

case "${MCP_STATELESS:-}" in
    true | 1) set -- "$@" --stateless ;;
esac

case "${MCP_ALLOW_UNAUTHENTICATED:-}" in
    true | 1) set -- "$@" --allow-unauthenticated ;;
esac

exec "$@"
