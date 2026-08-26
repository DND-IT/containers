# containers

Container images for cloud work, built multi-arch and rootless with security-first defaults. Inspired by [home-operations/containers](https://github.com/home-operations/containers).

**[Image size dashboard →](https://dnd-it.github.io/containers/)** — per-arch compressed sizes, layer breakdowns, and size history for every image, updated on each release.

## Images

| Image | Description |
|---|---|
| `ghcr.io/dnd-it/argocd-mcp` | [Argo CD MCP server](https://github.com/argoproj-labs/mcp-for-argocd) (applications, sync, and resources over MCP) |
| `ghcr.io/dnd-it/python` | Hardened Debian-slim Python base image for building and running Python services |
| `ghcr.io/dnd-it/go` | Hardened Debian Go toolchain base image for building Go services |
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

The server starts without credentials — a token is required per connection or per call, not at startup. Override the command for the other transports or to tune the listener:

```shell
# stdio, for clients that speak MCP over stdin/stdout
docker run --rm -i ghcr.io/dnd-it/argocd-mcp:latest argocd-mcp stdio

# more than one replica without sticky sessions needs stateless mode
docker run --rm -p 3000:3000 ghcr.io/dnd-it/argocd-mcp:latest argocd-mcp http --stateless
```

A `sse` transport also exists upstream but the HTTP Stream transport supersedes it.

### Verifying provenance

```shell
gh attestation verify oci://ghcr.io/dnd-it/argocd-mcp:0.8.0-1 --owner DND-IT
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
