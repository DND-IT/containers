"""Serve awslabs' AgentCore MCP server over streamable HTTP with only its browser tools.

The package's own entry point speaks stdio and registers every AgentCore service, including
the control-plane tools that create and delete runtimes, gateways and memories. Tools are
registered when the module is imported, so the allowlist has to be in the environment before
the import. Stateful HTTP, because a Browser session lives in this process: the Runtime keeps
a caller on one microVM for as long as its Mcp-Session-Id lasts.
"""

import os


def main() -> None:
    os.environ["AGENTCORE_ENABLE_TOOLS"] = "browser"

    import uvicorn
    from awslabs.amazon_bedrock_agentcore_mcp_server import server

    server.cache.ensure_ready()
    app = server.mcp.streamable_http_app(host="0.0.0.0", stateless_http=False)
    uvicorn.run(
        app,
        host=os.environ.get("MCP_HOST", "0.0.0.0"),
        port=int(os.environ.get("MCP_PORT", "8000")),
    )


if __name__ == "__main__":
    main()
