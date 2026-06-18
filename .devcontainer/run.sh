#!/usr/bin/env bash
# Host-side orchestration: prepare secrets, build the image, launch the autonomous
# build loop detached inside the container. Run from the host.
#
#   ./.devcontainer/run.sh                 # build command, primary model (GLM-5.2)
#   MAX_ITERS=50 ./.devcontainer/run.sh    # cap iterations for a short slice
#   MODEL=... ./.devcontainer/run.sh       # override the primary model
#
# Switch targets by editing .devcontainer/TASK.md, then re-running this script (it
# RECREATES the container — NEVER `docker start` a stopped one; that resurrects the old
# loop with stale task env and causes dual-loop API contention).
set -euo pipefail

REPO="/Volumes/m2-extended-disk/Repos/go-slides-mcp"
DC="$REPO/.devcontainer"
IMAGE="go-slides-builder"
NAME="go-slides-builder"
COMMAND="${1:-build}"
# PRIMARY: GLM-5.2 (free window) via HF Router -> Fireworks. loop.sh auto-falls-back to
# NVIDIA NIM Nemotron on a rate-limit signal. Override MODEL to change the primary.
MODEL="${MODEL:-huggingface/zai-org/GLM-5.2:fireworks-ai}"
VARIANT="${VARIANT-}"

echo "[run] preparing secrets..."
mkdir -p "$DC/secrets" "$DC/var"
gh auth token > "$DC/secrets/gh_token"
cp ~/.local/share/opencode/auth.json "$DC/secrets/auth.json"
chmod 600 "$DC/secrets/gh_token" "$DC/secrets/auth.json"
# Model keys: env override wins; otherwise keep the existing gitignored secret file.
# NEVER hardcode a token here — tokens live only in the gitignored secrets/ dir.
if [ -n "${HF_TOKEN:-}" ]; then printf '%s' "$HF_TOKEN" > "$DC/secrets/hf_token"; fi
if [ -f "$DC/secrets/hf_token" ]; then chmod 600 "$DC/secrets/hf_token"; else echo "[run] WARN: no secrets/hf_token — GLM-5.2 calls will fail"; fi
if [ -n "${NVIDIA_API_KEY:-}" ]; then printf '%s' "$NVIDIA_API_KEY" > "$DC/secrets/nvidia_api_key"; fi
if [ -f "$DC/secrets/nvidia_api_key" ]; then chmod 600 "$DC/secrets/nvidia_api_key"; else echo "[run] WARN: no secrets/nvidia_api_key — NIM fallback will fail"; fi
chmod +x "$DC/loop.sh" "$DC/entrypoint.sh"

echo "[run] building image (first build pulls Go 1.26 + Node + opencode + tooling)..."
docker build -t "$IMAGE" "$DC"

echo "[run] (re)starting container (recreate, never docker-start — avoids dual-loop contention)..."
docker rm -f "$NAME" >/dev/null 2>&1 || true
# The loop is PID 1. --restart on-failure brings it back on a crash but stays stopped after
# a clean [goal:complete]. Host ports: 8081->server 8080, 7101->dockyard inspector 7100
# (8080/7100 on the host may be taken by another project's builder).
docker run -d --name "$NAME" \
  --restart on-failure:20 \
  -p 127.0.0.1:8081:8080 \
  -p 127.0.0.1:7101:7100 \
  -v "$REPO:/workspace/go-slides-mcp" \
  -v "$DC/secrets:/run/secrets:ro" \
  -v "$DC/var:/var/go-slides" \
  -e COMMAND="$COMMAND" \
  -e MODEL="$MODEL" \
  -e VARIANT="$VARIANT" \
  -e MAX_ITERS="${MAX_ITERS:-2000}" \
  "$IMAGE" \
  bash -lc '/workspace/go-slides-mcp/.devcontainer/loop.sh'

# Clear any stale terminal status from a prior run BEFORE the new loop writes RUNNING
# (the entrypoint bootstrap delays loop.sh ~30-60s; without this, a monitor would catch the
# previous run's COMPLETE/BLOCKED in the bind-mounted var/ and false-fire).
mkdir -p "$DC/var"; echo "STARTING" > "$DC/var/status.txt"

# Inject the pptx-go engine skills + the Svelte/bridge study skill into the container's
# opencode skill dir (they live in host ~/.claude/skills, not the Go module cache).
for s in compose-a-scene define-a-theme scaffold-a-presentation load-a-brand-template \
         register-an-asset embed-a-chart-raster embed-a-code-block-raster extend-the-icon-set \
         dockyard-study-mcp; do
  docker cp ~/.claude/skills/"$s" "$NAME":/root/.config/opencode/skills/"$s" >/dev/null 2>&1 \
    && echo "[run] skill installed: $s" || echo "[run] WARN: skill copy failed: $s"
done

echo "[run] container up (command=$COMMAND, model=$MODEL). Bootstrap (secrets, dockyard, skills) runs before the loop."
echo "[run] monitor:  tail -f $DC/var/run.log   |   cat $DC/var/status.txt"
echo "[run] keep the host awake for long runs:  caffeinate -dimsu &"
