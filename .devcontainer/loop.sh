#!/usr/bin/env bash
# Outer autonomous build loop. Each iteration is a FRESH stateless opencode session that
# re-orients from git state + the orchestrator's brief (.devcontainer/TASK.md), so context
# stays lean across a long build (the git history IS the durable progress).
# Primary model: GLM-5.2 (free window); fallback: NVIDIA NIM Nemotron.
set -uo pipefail

# Login shells reset PATH; ensure all toolchain bins are reachable.
export PATH="/usr/local/bun/bin:/root/.opencode/bin:/usr/local/go/bin:/go/bin:/usr/local/bin:$PATH"

# Load the repo's opencode config (provider blocks + output limit + permission:allow +
# instructions). It lives under .devcontainer/, not the repo root, so point at it explicitly.
export OPENCODE_CONFIG="/workspace/go-slides-mcp/.devcontainer/opencode.json"

cd /workspace/go-slides-mcp

PRIMARY_MODEL="${MODEL:-openai/gpt-5.4}"          # GPT-5.4 low-effort via OpenAI oauth (opencode.json sets reasoningEffort)
PRIMARY_VARIANT="${VARIANT-}"   # empty VARIANT omits --variant
FALLBACK_MODEL="${FALLBACK_MODEL:-openai/gpt-5.4-mini}"   # lighter, same oauth account
FALLBACK_VARIANT="${FALLBACK_VARIANT-}"
ACTIVE="primary"
FB_ITERS=0
RETRY_PRIMARY_EVERY="${RETRY_PRIMARY_EVERY:-4}"
BUILD_PROMPT="$(cat /workspace/go-slides-mcp/.devcontainer/BUILD_PROMPT.md)"
BUILD_AGENT="${BUILD_AGENT:-build}"
MAX_ITERS="${MAX_ITERS:-2000}"
LOG="${LOG:-/var/go-slides/run.log}"
STATUS="/var/go-slides/status.txt"

mkdir -p "$(dirname "$LOG")"
note() { echo "$(date -u +%H:%M:%S) | $*" | tee -a "$LOG" >&2; }

note "=== build loop start (primary=$PRIMARY_MODEL fallback=$FALLBACK_MODEL agent=$BUILD_AGENT max=$MAX_ITERS) ==="
echo "RUNNING" > "$STATUS"

i=0
while [ "$i" -lt "$MAX_ITERS" ]; do
  i=$((i + 1))

  if [ "$ACTIVE" = "fallback" ] && [ "$FB_ITERS" -ge "$RETRY_PRIMARY_EVERY" ]; then
    note "retrying primary ($PRIMARY_MODEL) after $FB_ITERS fallback iterations"
    ACTIVE="primary"; FB_ITERS=0
  fi
  if [ "$ACTIVE" = "primary" ]; then
    CUR_MODEL="$PRIMARY_MODEL"; CUR_VARIANT="$PRIMARY_VARIANT"
  else
    CUR_MODEL="$FALLBACK_MODEL"; CUR_VARIANT="$FALLBACK_VARIANT"; FB_ITERS=$((FB_ITERS + 1))
  fi

  note "--- iteration $i (model=$CUR_MODEL${CUR_VARIANT:+ variant=$CUR_VARIANT}) ---"

  variant_args=()
  [ -n "$CUR_VARIANT" ] && variant_args=(--variant "$CUR_VARIANT")
  out="$(opencode run --model "$CUR_MODEL" "${variant_args[@]}" --agent "$BUILD_AGENT" "$BUILD_PROMPT" 2>&1)"
  printf '%s\n' "$out" >> "$LOG"

  if printf '%s' "$out" | grep -q '\[goal:complete\]'; then
    note "COMPLETE — unit reported done at iteration $i"; echo "COMPLETE" > "$STATUS"; break
  fi
  if printf '%s' "$out" | grep -q '\[goal:blocked\]'; then
    note "BLOCKED — model reported a blocker at iteration $i"; echo "BLOCKED" > "$STATUS"; break
  fi
  if printf '%s' "$out" | grep -qiE 'rate.?limit|quota exceeded|usage limit|429|402|insufficient[_ ]?(quota|credit|fund)|payment required|depleted|included credits|out of credits|purchase.*credits'; then
    if [ "$ACTIVE" = "primary" ]; then
      note "LIMIT on primary $CUR_MODEL (iter $i) → switching to fallback $FALLBACK_MODEL"
      echo "RATE_LIMITED->fallback (iter $i)" > "$STATUS"; ACTIVE="fallback"; FB_ITERS=0
    else
      note "LIMIT on fallback $CUR_MODEL (iter $i) — backing off 120s"
      echo "RATE_LIMITED (iter $i)" > "$STATUS"; sleep 120
    fi
  fi

  sleep 5
done

note "=== loop ended (iterations=$i) ==="
if [ -f "$STATUS" ] && [ "$(cat "$STATUS")" = "RUNNING" ]; then
  echo "MAX_ITERS_REACHED" > "$STATUS"
fi
# Always exit 0 on a clean end so the restart policy does NOT relaunch after [goal:complete].
exit 0
