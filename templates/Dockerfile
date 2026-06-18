# Autonomous-build container.
# Toolchain ONLY (language runtime, package managers, the framework CLI, git/gh, opencode).
# No secrets are baked in — they arrive at runtime via the read-only /run/secrets mount.
# ADAPT: the base image + toolchain to your project (this example targets a Go + Node build).
FROM golang:1.25-bookworm

ENV DEBIAN_FRONTEND=noninteractive

# Base tooling. (If your build needs a C toolchain, add build-essential; a CGO_ENABLED=0
# Go build does not.) `make` is here for Makefile-driven gates.
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl git unzip ripgrep jq make \
    && rm -rf /var/lib/apt/lists/*

# Node.js (only if the project has a JS/TS UI to build).
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

# Linter PINNED to the CI version so an in-container `lint` matches the merge gate exactly.
# `go install` it (not a prebuilt binary): golangci-lint refuses to run when built with an
# older Go than the go.mod target, so it must compile with THIS image's toolchain. Clean the
# ~750MB build cache in the same RUN so only the ~50MB binary remains.
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 \
    && golangci-lint version \
    && go clean -cache

# Bun (opencode plugin installer) + opencode CLI (the builder agent).
ENV BUN_INSTALL=/usr/local/bun
ENV PATH=$BUN_INSTALL/bin:/root/.opencode/bin:$PATH
RUN curl -fsSL https://bun.sh/install | bash
RUN curl -fsSL https://opencode.ai/install | bash

# Toolchain bins on PATH for login/interactive shells too — the loop runs `bash -lc`,
# which sources /etc/profile.d and would otherwise reset PATH and lose the toolchain.
RUN printf 'export PATH="/usr/local/bun/bin:/root/.opencode/bin:/usr/local/go/bin:/go/bin:/usr/local/bin:$PATH"\n' \
        > /etc/profile.d/toolchain.sh

WORKDIR /workspace/project

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["sleep", "infinity"]
