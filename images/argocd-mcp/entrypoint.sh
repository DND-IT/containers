#!/bin/sh
# Upstream exposes the port, stateless and allow-unauthenticated settings as
# flags only. Translate them from the environment so the image runs unchanged
# on platforms that cannot override the container command.
set -eu

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
