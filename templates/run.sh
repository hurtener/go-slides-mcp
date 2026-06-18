#!/usr/bin/env bash
# Host-side orchestration: prepare secrets, build the image, and launch the
# plan-driven autonomous build loop detached inside the container. Run from the host.
#
#   ./.devcontainer/run.sh                 # default: build command, primary model
#   MAX_ITERS=50 ./.devcontainer/run.sh    # cap iterations for a short slice
#   MODEL=... ./.devcontainer/run.sh       # override the primary model
#
# Switch targets by editing .devcontainer/TASK.md, then re-running this script
# (it RECREATES the container — NEVER `docker start` a stopped one; that resurrects
# the old loop with stale task env and causes dual-loop API contention).
#
# ADAPT: set REPO/IMAGE/NAME and the bind-mount target path for your project.
set -euo pipefail

REPO="/ABSOLUTE/PATH/TO/your-project"      # <-- set to the new project's repo root
DC="$REPO/.devcontainer"
IMAGE="myproject-builder"                   # <-- name the image
NAME="myproject-builder"                    # <-- name the container
COMMAND="${1:-build}"
# PRIMARY model: GLM-5.2 (free window) via the HF Router → Fireworks provider.
# After the free window, fall back to NVIDIA NIM Nemotron (see AUTONOMOUS_BUILD_LOOP.md §5).
MODEL="${MODEL:-huggingface/zai-org/GLM-5.2:fireworks-ai}"
VARIANT="${VARIANT-}"

echo "[run] preparing secrets..."
mkdir -p "$DC/secrets" "$DC/var"
gh auth token > "$DC/secrets/gh_token"
cp ~/.local/share/opencode/auth.json "$DC/secrets/auth.json"
chmod 600 "$DC/secrets/gh_token" "$DC/secrets/auth.json"
# Model API keys: env override wins; otherwise keep the existing gitignored secret file.
# NEVER hardcode a token here — tokens live only in the gitignored secrets/ dir.
if [ -n "${HF_TOKEN:-}" ]; then printf '%s' "$HF_TOKEN" > "$DC/secrets/hf_token"; fi
if [ -f "$DC/secrets/hf_token" ]; then chmod 600 "$DC/secrets/hf_token"; else echo "[run] WARN: no secrets/hf_token — GLM-5.2 calls will fail"; fi
if [ -n "${NVIDIA_API_KEY:-}" ]; then printf '%s' "$NVIDIA_API_KEY" > "$DC/secrets/nvidia_api_key"; fi
if [ -f "$DC/secrets/nvidia_api_key" ]; then chmod 600 "$DC/secrets/nvidia_api_key"; else echo "[run] WARN: no secrets/nvidia_api_key — NIM fallback will fail"; fi
chmod +x "$DC/loop.sh" "$DC/entrypoint.sh"

echo "[run] building image (first build pulls the toolchain + opencode)..."
docker build -t "$IMAGE" "$DC"

echo "[run] (re)starting container (recreate, never docker-start — avoids dual-loop contention)..."
docker rm -f "$NAME" >/dev/null 2>&1 || true
# The loop is PID 1. --restart on-failure brings it back on a crash (OOM, daemon hiccup),
# but a clean [goal:complete] exits 0 so it stays stopped. Publish any dev port you need
# for orchestrator-side live validation (e.g. -p 127.0.0.1:8080:8080).
docker run -d --name "$NAME" \
  --restart on-failure:20 \
  -p 127.0.0.1:8080:8080 \
  -v "$REPO:/workspace/project" \
  -v "$DC/secrets:/run/secrets:ro" \
  -v "$DC/var:/var/project" \
  -e COMMAND="$COMMAND" \
  -e MODEL="$MODEL" \
  -e VARIANT="$VARIANT" \
  -e MAX_ITERS="${MAX_ITERS:-2000}" \
  "$IMAGE" \
  bash -lc '/workspace/project/.devcontainer/loop.sh'

echo "[run] container up (command=$COMMAND, model=$MODEL). Bootstrap runs before the loop."
echo "[run] monitor:  tail -f $DC/var/run.log   |   cat $DC/var/status.txt"
echo "[run] keep the host awake for long runs:  caffeinate -dimsu &"
