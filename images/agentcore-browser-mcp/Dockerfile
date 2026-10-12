FROM ghcr.io/dnd-it/python:0.1.0@sha256:b633d0deb7898b22c6737438d0458c93669bdd8bd24dc5717e6663cdaab83731

# renovate: datasource=pypi depName=awslabs.amazon-bedrock-agentcore-mcp-server
ARG AGENTCORE_MCP_VERSION=0.2.1
# renovate: datasource=pypi depName=uvicorn
ARG UVICORN_VERSION=0.54.0

ENV PATH="/home/python/.local/bin:${PATH}" \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

RUN pip3 install --user --no-cache-dir \
        "awslabs.amazon-bedrock-agentcore-mcp-server==${AGENTCORE_MCP_VERSION}" \
        "uvicorn==${UVICORN_VERSION}" \
    && python3 -c "import awslabs.amazon_bedrock_agentcore_mcp_server, uvicorn"

COPY --chmod=755 serve.py /opt/agentcore-browser-mcp/serve.py

LABEL org.opencontainers.image.description="AgentCore MCP server with only its AgentCore Browser tools, over streamable HTTP"

EXPOSE 8000

ENTRYPOINT ["python3", "/opt/agentcore-browser-mcp/serve.py"]
