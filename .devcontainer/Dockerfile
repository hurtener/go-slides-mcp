# go-slides-mcp (Deckard) autonomous-build container.
# Toolchain ONLY (Go 1.26 + Node 22 + gh + Bun + opencode). Dockyard is installed at
# RUNTIME by entrypoint.sh (it is a private github.com/hurtener module that needs the
# gh token, which never gets baked into the image). No secrets are baked in.
FROM golang:1.26-bookworm

ENV DEBIAN_FRONTEND=noninteractive

# Base tooling. CGO_ENABLED=0 Go builds need no C toolchain; `make` drives the gates.
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl git unzip ripgrep jq make \
    && rm -rf /var/lib/apt/lists/*

# Node.js 22 LTS + npm — for the Svelte/Vite MCP App UIs (the 3 ui:// surfaces).
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm --version && node --version

# GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
        | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
        > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

# golangci-lint PINNED to the host/CI version (v2 line — v1.x refuses to run under a
# Go 1.26 go.mod). `go install` it (not a prebuilt binary) so it compiles with THIS
# image's Go 1.26 toolchain. Clean the build cache in the same RUN so only the binary remains.
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 \
    && golangci-lint version \
    && go clean -cache

# Bun (opencode plugin installer) + opencode CLI (the builder agent).
ENV BUN_INSTALL=/usr/local/bun
ENV PATH=$BUN_INSTALL/bin:/root/.opencode/bin:$PATH
RUN curl -fsSL https://bun.sh/install | bash
RUN curl -fsSL https://opencode.ai/install | bash

# Toolchain bins on PATH for login/interactive shells too — the loop runs `bash -lc`,
# which sources /etc/profile.d and would otherwise reset PATH and lose the toolchain.
# GOPRIVATE so `go install`/`go get` of github.com/hurtener/{dockyard,pptx-go} bypass the proxy.
RUN printf 'export PATH="/usr/local/bun/bin:/root/.opencode/bin:/usr/local/go/bin:/go/bin:/usr/local/bin:$PATH"\nexport GOPRIVATE=github.com/hurtener\n' \
        > /etc/profile.d/toolchain.sh
ENV GOPRIVATE=github.com/hurtener

WORKDIR /workspace/go-slides-mcp

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["sleep", "infinity"]
