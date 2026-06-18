#!/usr/bin/env bash
# Runtime bootstrap for the go-slides-mcp (Deckard) build loop: inject secrets,
# configure git/gh/go, install Dockyard + its skills. Secrets arrive read-only at
# /run/secrets (bind mount) — never baked into the image.
set -uo pipefail
SEC=/run/secrets

echo "[entrypoint] bootstrapping $(date -u)"

# 1. opencode auth — the OAuth that authorizes opencode (+ any OAuth providers).
if [ -f "$SEC/auth.json" ]; then
  mkdir -p /root/.local/share/opencode
  cp "$SEC/auth.json" /root/.local/share/opencode/auth.json
  chmod 600 /root/.local/share/opencode/auth.json
  echo "[entrypoint] opencode auth installed"
else
  echo "[entrypoint] WARN: no auth.json in /run/secrets — model calls may fail"
fi

# 2a. HF_TOKEN — authorizes GLM-5.2 via the HF Router -> Fireworks provider (PRIMARY).
#     Exported for `exec "$@"` AND written to profile.d (the loop runs as `bash -lc`).
if [ -f "$SEC/hf_token" ]; then
  HF_TOKEN="$(tr -d '[:space:]' < "$SEC/hf_token")"; export HF_TOKEN
  echo "export HF_TOKEN=$HF_TOKEN" > /etc/profile.d/hf.sh; chmod 644 /etc/profile.d/hf.sh
  echo "[entrypoint] HF token installed (HF_TOKEN set)"
else
  echo "[entrypoint] WARN: no hf_token in /run/secrets — GLM-5.2 calls will fail"
fi

# 2b. NVIDIA NIM API key — authorizes the OpenAI-compatible NIM provider (FALLBACK).
if [ -f "$SEC/nvidia_api_key" ]; then
  NVIDIA_API_KEY="$(tr -d '[:space:]' < "$SEC/nvidia_api_key")"; export NVIDIA_API_KEY
  echo "export NVIDIA_API_KEY=$NVIDIA_API_KEY" > /etc/profile.d/nvidia.sh; chmod 644 /etc/profile.d/nvidia.sh
  echo "[entrypoint] NVIDIA NIM key installed (NVIDIA_API_KEY set)"
else
  echo "[entrypoint] WARN: no nvidia_api_key in /run/secrets — NIM fallback will fail"
fi

# 3. GitHub token — gh CLI (PRs) + git over HTTPS + private Go module access.
if [ -f "$SEC/gh_token" ]; then
  TOKEN="$(tr -d '[:space:]' < "$SEC/gh_token")"
  echo "$TOKEN" | gh auth login --with-token >/dev/null 2>&1 \
    && echo "[entrypoint] gh authenticated" \
    || echo "[entrypoint] WARN: gh auth login failed"
  git config --global url."https://x-access-token:${TOKEN}@github.com/".insteadOf "https://github.com/"
else
  echo "[entrypoint] WARN: no gh_token — pushes and private module fetches will fail"
fi

# 3b. Point opencode at the repo config (.devcontainer/opencode.json is not at the repo
#     root, so opencode won't auto-load it from cwd) for every login shell.
echo 'export OPENCODE_CONFIG=/workspace/go-slides-mcp/.devcontainer/opencode.json' > /etc/profile.d/opencode-config.sh
chmod 644 /etc/profile.d/opencode-config.sh

# 4. git identity — PERSONAL hurtener account, signing OFF (never the work account).
#    (The builder must not commit anyway — the orchestrator owns all git — but set it safely.)
git config --global user.name "Santi Benvenuto"
git config --global user.email "117486687+hurtener@users.noreply.github.com"
git config --global commit.gpgsign false
git config --global tag.gpgsign false
git config --global init.defaultBranch main
git config --global --add safe.directory /workspace/go-slides-mcp

# 5. Dockyard CLI from source (private repo, needs token from step 3) — retry on transient failure.
if ! command -v dockyard >/dev/null 2>&1; then
  echo "[entrypoint] installing dockyard from source..."
  for attempt in 1 2 3 4 5; do
    if GOPRIVATE=github.com/hurtener GOBIN=/usr/local/bin \
        go install github.com/hurtener/dockyard/cmd/dockyard@v1.7.3; then
      echo "[entrypoint] dockyard installed: $(dockyard --version 2>&1 | head -1)"
      break
    fi
    echo "[entrypoint] dockyard install attempt $attempt failed; retrying in 10s..."
    sleep 10
  done
  command -v dockyard >/dev/null 2>&1 || echo "[entrypoint] WARN: dockyard still missing after retries"
fi

# 6. Dockyard Agent Skills — copy from the Go module cache into opencode's global skills
#    dir so the build agent can use them via the `skill` tool. (pptx-go skills are injected
#    by run.sh via `docker cp`, since they live in the host ~/.claude, not this module cache.)
SKILLS_SRC="$(ls -d /go/pkg/mod/github.com/hurtener/dockyard@*/skills 2>/dev/null | head -1)"
if [ -n "$SKILLS_SRC" ]; then
  mkdir -p /root/.config/opencode/skills
  cp -R "$SKILLS_SRC"/* /root/.config/opencode/skills/ 2>/dev/null || true
  chmod -R u+w /root/.config/opencode/skills 2>/dev/null || true
  find /root/.config/opencode/skills -name .DS_Store -delete 2>/dev/null || true
  echo "[entrypoint] dockyard skills installed: $(ls /root/.config/opencode/skills 2>/dev/null | tr '\n' ' ')"
else
  echo "[entrypoint] WARN: dockyard skills not found in module cache yet"
fi

# 7. Warm the module cache (best-effort; never fail bootstrap on a transient hiccup).
( cd /workspace/go-slides-mcp && go mod download 2>/dev/null ) || true

echo "[entrypoint] ready"
exec "$@"
