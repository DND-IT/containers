# containers

Container images for cloud work, built multi-arch and rootless with security-first defaults. Inspired by [home-operations/containers](https://github.com/home-operations/containers).

**[Image size dashboard →](https://dnd-it.github.io/containers/)** — per-arch compressed sizes, layer breakdowns, and size history for every image, updated on each release.

## Images

| Image | Description |
|---|---|
| `ghcr.io/dnd-it/argocd-mcp` | [Argo CD MCP server](https://github.com/argoproj-labs/mcp-for-argocd) (applications, sync, and resources over MCP) |
| `ghcr.io/dnd-it/agentcore-browser-mcp` | [AgentCore MCP server](https://pypi.org/project/awslabs.amazon-bedrock-agentcore-mcp-server/) with only its AgentCore Browser tools, over streamable HTTP |
| `ghcr.io/dnd-it/secret-env` | Entrypoint shim that exports an AWS Secrets Manager secret's strings and execs the server |
| `ghcr.io/dnd-it/github-mcp` | [GitHub's MCP server](https://github.com/github/github-mcp-server), read-only over streamable HTTP, with installation tokens from an octo-sts token broker |
| `ghcr.io/dnd-it/claude-code` | Headless [Claude Code](https://www.npmjs.com/package/@anthropic-ai/claude-code) (`claude -p`) with `git`, `gh`, `uv` and `rg`; the entrypoint can clone an agent profile into `CLAUDE_CONFIG_DIR` and a repository into `/workspace` first |
| `ghcr.io/dnd-it/python` | Hardened Debian-slim Python base image for building and running Python services |
| `ghcr.io/dnd-it/go` | Hardened Debian Go toolchain base image for building Go services |
| `ghcr.io/dnd-it/go-runtime` | Minimal Debian-slim runtime base for statically-compiled Go binaries, no toolchain |
| `ghcr.io/dnd-it/node` | Hardened Debian-slim Node.js base image for building and running Node.js and TypeScript services |

## Usage

Pin to a semver tag plus digest so tools like Renovate can track updates reliably:

```yaml
image: ghcr.io/dnd-it/<name>:1.0.0@sha256:<digest>
```

Every image is tagged `X.Y.Z`, `X.Y`, `X`, and `latest`. Images that track an
upstream release may carry a `-<revision>` suffix on the full tag (`2.336.0-1`)
when this repo rebuilds the same upstream version; the rolling `X.Y`, `X`, and
`latest` tags drop it.

### Defaults

- Multi-arch: `linux/amd64` and `linux/arm64`, each built on native runners (no QEMU)
- Rootless: processes run as a dedicated non-root user (uid `1001`), no sudo
- One process per container, logs to stdout, no init frameworks
- No docker CLI, except CI-runner images that build against an injected dind sidecar
- Base images pinned by digest, tool versions pinned and updated by Renovate
- SBOM and SLSA provenance attestations attached to every image
- Every published image re-scanned daily for HIGH/CRITICAL CVEs, reported to GitHub code scanning

### argocd-mcp

Packages the npm release of [argoproj-labs/mcp-for-argocd](https://github.com/argoproj-labs/mcp-for-argocd). Defaults to the HTTP Stream transport on `:3000` (`POST /mcp`, with `GET /healthz` for probes):

```shell
docker run --rm -p 3000:3000 \
  -e ARGOCD_BASE_URL=https://argocd.example.com \
  -e ARGOCD_API_TOKEN=<token> \
  ghcr.io/dnd-it/argocd-mcp:latest
```

| Variable | Effect |
|---|---|
| `ARGOCD_BASE_URL` | Argo CD API endpoint; may also be sent per-connection as the `x-argocd-base-url` header |
| `ARGOCD_API_TOKEN` | Argo CD API token; may also be sent as the `x-argocd-api-token` header. Never read from tool arguments |
| `ARGOCD_TOKEN_REGISTRY_PATH` | JSON file of `{baseUrl, token}` pairs for targeting multiple Argo CD instances. Fails closed if set but unreadable |
| `MCP_READ_ONLY` | `true` disables `create_application`, `update_application`, `delete_application`, and `sync_application` |

The server starts without credentials — a token is required per connection or per call, not at startup.

Upstream exposes the listener settings as flags only. This image's entrypoint also reads them from the environment, so it runs on platforms that cannot override the container command:

| Variable | Effect |
|---|---|
| `MCP_BIND_ADDRESS` | Address to listen on. Defaults to loopback; a wider bind needs `MCP_AUTH_TOKEN` or `MCP_ALLOW_UNAUTHENTICATED` |
| `MCP_PORT` | Listener port, default `3000` |
| `MCP_STATELESS` | `true` runs stateless, for more than one replica without sticky sessions |
| `MCP_ALLOW_UNAUTHENTICATED` | `true` permits a non-loopback bind with no `MCP_AUTH_TOKEN`. Only safe when an external layer already authenticates callers |
| `MCP_AUTH_TOKEN` | Bearer token required on inbound requests |

Runtime environment variables are not always secret, so the image starts through [`secret-env`](#secret-env), which can load credentials from AWS Secrets Manager before the server starts:

| Variable | Effect |
|---|---|
| `SECRET_ENV_ID` | Name or ARN of a JSON secret of string variables, e.g. `{"ARGOCD_API_TOKEN": "..."}`. Read with the default AWS credential chain. The container exits if the secret cannot be read or holds anything but strings |

Amazon Bedrock AgentCore Runtime expects streamable HTTP on `0.0.0.0:8000/mcp` and authenticates every invocation with IAM before it reaches the container, so it needs `MCP_BIND_ADDRESS=0.0.0.0`, `MCP_PORT=8000`, `MCP_STATELESS=true` and `MCP_ALLOW_UNAUTHENTICATED=true`.

Any command passed to the container runs verbatim instead, so the other transports and explicit flags still work:

```shell
# stdio, for clients that speak MCP over stdin/stdout
docker run --rm -i ghcr.io/dnd-it/argocd-mcp:latest argocd-mcp stdio

# more than one replica without sticky sessions needs stateless mode
docker run --rm -p 3000:3000 ghcr.io/dnd-it/argocd-mcp:latest argocd-mcp http --stateless
```

A `sse` transport also exists upstream but the HTTP Stream transport supersedes it.

### github-mcp

[GitHub's MCP server](https://github.com/github/github-mcp-server), built as a library into a
small Go program that serves it read-only over streamable HTTP on `0.0.0.0:8000/mcp`
(`MCP_HOST`, `MCP_PORT`). Upstream's HTTP mode takes the GitHub token from each request's
`Authorization` header, which a platform such as Amazon Bedrock AgentCore Runtime uses for its
own signature. So upstream listens on loopback, and the program in front of it sets a GitHub App
installation token on every request it forwards:

| Variable | Effect |
|---|---|
| `TOKEN_BROKER_DOMAIN` | The [octo-sts](https://github.com/octo-sts/app) broker's domain; also the audience of the STS web identity token the container's AWS role presents to it |
| `TOKEN_BROKER_SCOPE` | What the token is for: an organization (policy in its `.github` repository) or `owner/repo` |
| `TOKEN_BROKER_IDENTITY` | The trust policy's name, `.github/chainguard/<identity>.sts.yaml` |
| `GITHUB_TOOLSETS` | Comma-separated toolsets to enable; upstream's default set when unset |
| `GITHUB_PERSONAL_ACCESS_TOKEN` | For local runs: a fixed token instead of the broker |

The role needs `sts:GetWebIdentityToken` for the broker's domain as audience, and the account's
outbound web identity federation enabled. A token is bought at startup, which fails if the
broker refuses, and renewed after 45 minutes; if renewal fails the current one is kept until 55.

From MCP protocol `2026-07-28` upstream answers `400` to a `tools/call` whose `owner` or `repo`
argument has no matching `Mcp-Param-*` header. A client that does not send them, such as an
AgentCore Gateway, gets them from the program in front, copied from the request body; headers the
client sends are kept.

### secret-env

A static binary for other images to start their server through, so a server that takes its
credentials from the environment can run on a platform whose environment variables are not
secret, such as Amazon Bedrock AgentCore Runtime. Copy it in and put it in front of the command:

```dockerfile
COPY --from=ghcr.io/dnd-it/secret-env:0.1.0@sha256:<digest> /usr/local/bin/secret-env /usr/local/bin/secret-env
ENTRYPOINT ["/usr/local/bin/secret-env"]
CMD ["my-server"]
```

| Variable | Effect |
|---|---|
| `SECRET_ENV_ID` | Name or ARN of a Secrets Manager secret holding a JSON object of strings, e.g. `{"ARGOCD_API_TOKEN": "..."}`. Each one is exported before the command starts. Read with the default AWS credential chain; a name takes its region from `AWS_REGION` |

Without `SECRET_ENV_ID` the command runs untouched. With it, the command does not start if the
secret cannot be read, is not an object, or holds a value that is not a string.

### agentcore-browser-mcp

Packages awslabs' [AgentCore MCP server](https://pypi.org/project/awslabs.amazon-bedrock-agentcore-mcp-server/)
with every service but AgentCore Browser switched off, so the control-plane tools that create
and delete runtimes, gateways and memories never exist in the process. It serves stateful
streamable HTTP on `0.0.0.0:8000/mcp` (`MCP_HOST`, `MCP_PORT`), because each Browser session
lives in the process, and calls AgentCore with the default AWS credential chain; the role needs
the `bedrock-agentcore` browser session actions.

### claude-code

Headless [Claude Code](https://www.npmjs.com/package/@anthropic-ai/claude-code): the default command is `claude -p`, which reads the prompt from stdin, does the work in `/workspace` and prints the result. Before it runs, the entrypoint can put an agent profile (`CLAUDE.md`, `agents/`, `skills/`, `settings.json`) into `CLAUDE_CONFIG_DIR` and clone a repository to work on:

```shell
docker run --rm \
  -e CLAUDE_CODE_USE_BEDROCK=1 -e AWS_REGION=eu-central-1 \
  -e AGENT_PROFILE_REPO=https://github.com/DND-IT/<profiles> -e AGENT_PROFILE=reviewer \
  -e AGENT_REPO=https://github.com/DND-IT/<repo> -e GITHUB_TOKEN=<token> \
  -e AGENT_PROMPT="Review the last commit" \
  ghcr.io/dnd-it/claude-code:latest
```

| Variable | Effect |
|---|---|
| `AGENT_PROFILE_REPO` | Git URL whose contents (or `AGENT_PROFILE` within it) are copied into `CLAUDE_CONFIG_DIR`; `AGENT_PROFILE_REF` picks a branch or tag |
| `AGENT_REPO` | Git URL cloned into `/workspace` unless it already holds a checkout; `AGENT_REPO_REF` picks a branch |
| `AGENT_PROMPT` | Sent to the command on stdin, so `claude -p` runs it without a pipe |
| `GITHUB_TOKEN` | Used for `https://github.com` through a git credential helper, so it is never written to `~/.gitconfig` or a remote URL |
| `CLAUDE_CONFIG_DIR` | Where Claude Code reads its settings, agents and skills; `/home/node/.claude` |

Authentication is Claude Code's own: `ANTHROPIC_API_KEY`, or `CLAUDE_CODE_USE_BEDROCK=1` with AWS credentials. Any other command runs after the same bootstrap, e.g. `... ghcr.io/dnd-it/claude-code:latest bash`.

### Verifying provenance

```shell
gh attestation verify oci://ghcr.io/dnd-it/argocd-mcp:0.9.0-1 --owner DND-IT
```

## Versioning and releases

Each image is versioned independently via the `version` field in its `images/<name>/metadata.yaml`:

- Images that package a single upstream application (e.g. `actions-runner`) track the upstream version, with an optional `-<revision>` suffix for rebuilds that change the image without moving the upstream version.
- Images owned by this repo use their own semver: MAJOR for breaking changes (removed tools, changed users/paths), MINOR for additions, PATCH for fixes and rebuilds.

CI builds and tags whatever version the metadata declares — bump it in the same PR as the change.

## Adding an image

1. Create `images/<name>/Dockerfile` and `images/<name>/metadata.yaml`.
2. Pin the base image by digest and any downloaded tools with a `# renovate:` annotation.
3. Create a non-root user (uid `1001`) and switch to it with `USER`.
4. Add a `test` command to the metadata — CI runs it against the built image on PRs.

The `add-image` skill (`.claude/skills/add-image/SKILL.md`) walks through all of it.

## CI

`.github/workflows/build.yaml` builds only images changed in a PR or push:

1. **prepare** — diffs `images/` and emits a build matrix from each image's `metadata.yaml`.
2. **build** — one job per image per platform, on native runners (`ubuntu-latest` for amd64, `ubuntu-24.04-arm` for arm64) with BuildKit and GitHub Actions layer caching. PRs build and smoke-test locally; pushes to `main` push by digest with SBOM and provenance.
3. **merge** — stitches the per-arch digests into one manifest list, applies the semver tags, and attests build provenance.
4. **sizes** — collects per-arch sizes for the published tags and publishes them to the `gh-pages` dashboard.

Trigger a manual build of any (or every) image via *Actions → Build → Run workflow*.

`.github/workflows/scan.yaml` runs daily and scans every published image, one job per image per platform on its native runner, uploading Trivy results to GitHub code scanning. A CVE disclosed after an image ships only surfaces on a re-scan, so this — not the build — is what catches them.

The scan reports rather than gates. Almost everything it finds lives in vendored upstream artefacts (Go binaries, .NET runtimes, npm's own bundled dependencies) where a patched upstream module exists but this repo only consumes a release build, so it cannot act on the fix until upstream rebuilds. Gating merges on that would block PRs on work the repo cannot do. Treat the alerts as a queue: the ones worth acting on are the OS packages a base-image bump fixes, and language deps this repo installs directly.

### Registry namespace

The org login is `DND-IT`, but OCI references must be lowercase, so workflows use a
`NAMESPACE: dnd-it` env var instead of `github.repository_owner`.

## Local development

Build with [Apple container](https://github.com/apple/container) (or any BuildKit-compatible builder):

```shell
container build -t <name>:local -f images/<name>/Dockerfile images/<name>
container run --rm <name>:local <test command from metadata.yaml>
```
