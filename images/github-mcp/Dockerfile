FROM ghcr.io/dnd-it/go:0.1.0@sha256:130495cac03114ff8a4b6f1cb45d2706ee504ccd8f9754395fd59ea8b9873595 AS build

WORKDIR /src
COPY --chown=go:go go.mod go.sum ./
RUN go mod download
COPY --chown=go:go *.go ./
RUN go vet ./... \
    && go test ./... \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /tmp/github-mcp .

FROM ghcr.io/dnd-it/go-runtime:0.1.0@sha256:5cbcb258cf8253667ba6099e083e8a543f1561c389ca1d1bbac39cbe5d15cb93

COPY --from=build /tmp/github-mcp /usr/local/bin/github-mcp

LABEL org.opencontainers.image.description="GitHub's MCP server, read-only over streamable HTTP, with installation tokens from an octo-sts token broker"

EXPOSE 8000

ENTRYPOINT ["/usr/local/bin/github-mcp"]
