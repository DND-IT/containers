#!/usr/bin/env bash
set -euo pipefail

workspace="${AGENT_WORKSPACE:-/workspace}"
profile_dir="${CLAUDE_CONFIG_DIR:-${HOME}/.claude}"

# The token stays in the environment: git asks this helper for it on every
# fetch, so it is never written to ~/.gitconfig or a remote URL.
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    git config --global credential.https://github.com.helper \
        '!f() { echo username=x-access-token; echo "password=${GITHUB_TOKEN}"; }; f'
fi

if [[ -n "${AGENT_PROFILE_REPO:-}" ]]; then
    checkout="$(mktemp -d)"
    git clone --quiet --depth 1 ${AGENT_PROFILE_REF:+--branch "${AGENT_PROFILE_REF}"} \
        "${AGENT_PROFILE_REPO}" "${checkout}"
    source="${checkout}${AGENT_PROFILE:+/${AGENT_PROFILE}}"
    if [[ ! -d "${source}" ]]; then
        echo "agent profile not found in ${AGENT_PROFILE_REPO}: ${AGENT_PROFILE:-<root>}" >&2
        exit 1
    fi
    rm -rf "${checkout}/.git"
    mkdir -p "${profile_dir}"
    cp -aT "${source}" "${profile_dir}"
    rm -rf "${checkout}"
fi

mkdir -p "${workspace}"
if [[ -n "${AGENT_REPO:-}" && ! -d "${workspace}/.git" ]]; then
    git clone --quiet ${AGENT_REPO_REF:+--branch "${AGENT_REPO_REF}"} "${AGENT_REPO}" "${workspace}"
fi
cd "${workspace}"

if [[ -n "${AGENT_PROMPT:-}" ]]; then
    exec "$@" <<<"${AGENT_PROMPT}"
fi

exec "$@"
