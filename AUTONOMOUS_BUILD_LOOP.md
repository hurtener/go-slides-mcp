# The Autonomous Build Loop

A reusable methodology for building a large, well-specified software project **fast and
cheaply** by delegating the mechanical implementation volume to a cheap/free "builder"
model running unattended in a container, while a capable "orchestrator" model (you —
e.g. Claude) plans the work, owns verification, and gates quality.

> You are the **architect and the QA**, not the typist. The builder writes most of the
> lines; you decide what gets built, you prove it actually works, and you own the bar.

This document is project-agnostic. It is distilled from a reference implementation
(a Go MCP gateway) and from the Claude Code skill **`orchestrate-autonomous-build`**,
which is the canonical write-up of this method. Use this file to stand the loop up on a
new project (e.g. `go-slides-mcp`) in ~15 minutes. Sanitized copies of every config file
live in [`templates/`](./templates/).

---

## Quickstart (TL;DR)

```bash
# 0. One-time: copy the templates into the new repo and adapt them (see §8).
cp -r templates .devcontainer            # then edit paths/model/gates inside

# 1. Orchestrator: write ONE coherent unit of work into .devcontainer/TASK.md
#    (gitignored scratch; the per-iteration spec for the builder).

# 2. Launch the loop (builds the Docker image on first run, then runs unattended):
./.devcontainer/run.sh
caffeinate -dimsu &                       # macOS: keep the host awake for long runs

# 3. Monitor on a timer (~every 20-30 min — don't hover):
tail -f .devcontainer/var/run.log
cat .devcontainer/var/status.txt          # RUNNING | COMPLETE | BLOCKED | RATE_LIMITED | MAX_ITERS_REACHED
git -C . status && git -C . log --oneline -5

# 4. On [goal:complete]: DO NOT trust it. Independently verify (gates + read the code +
#    run it live), fix the hard parts yourself, then commit on a branch, PR, merge.

# 5. Switch to the next unit by editing TASK.md and re-running run.sh (it RECREATES
#    the container — never `docker start` a stopped one).
```

---

## 1. Concept & when to use it

The method is a **hybrid model split**:

- **Cheap/free builder model** does breadth and mechanical volume — scaffolding similar
  units, filling boilerplate, replicating an established pattern across N modules,
  cloning storage/handlers/CLI/pure-function code.
- **You, the capable orchestrator,** do the work that is novel, cross-cutting, or
  high-judgment: the plan, the first reference implementation, the hard slices, and —
  above all — **independent verification and quality gating.**

The single biggest failure mode is letting the builder's *claim* of completion stand in
for *your* verification (see §7). The builder's `[goal:complete]` is where verification
**starts**, never where it ends.

**Use this when** the project is large, well-specified, and decomposable into many
similar units against machine-checkable gates (build/test/lint/validate). It shines on
phase-driven work with binding acceptance criteria. **Don't use it** for small one-shot
tasks, or for work that is mostly novel/cross-cutting with no clonable pattern (just do
that yourself).

> **Canonical reference:** the Claude Code skill **`orchestrate-autonomous-build`** —
> "delegate heavy implementation to cheap/free models running unattended in a container
> while a capable model (you) plans the work, owns verification, and gates quality."
> Read it before your first run; this doc is the operational distillation.

**The shape, in one line:** orchestrator plan → builder implement → builder iterate →
orchestrator adversary → builder fix → orchestrator evaluate + final fix → orchestrator
commit + PR + merge → repeat.

---

## 2. Architecture

```
HOST (your machine, capable model = orchestrator)
 │
 │  ./.devcontainer/run.sh
 │    1. writes secrets/   (gh token, opencode auth.json, model API key) — gitignored
 │    2. docker build  →  builder image (toolchain only, NO secrets baked in)
 │    3. docker rm -f old container; docker run -d new one
 │
 ▼
CONTAINER (cheap/free model = builder)
 │  entrypoint.sh  → injects secrets from /run/secrets (read-only bind mount),
 │                    sets git identity, points OPENCODE_CONFIG at the repo config
 │  loop.sh  (PID 1)
 │    while not done:
 │      opencode run --agent build "<BUILD_PROMPT>"   # FRESH, stateless each iteration
 │      grep output for [goal:complete] / [goal:blocked] / rate-limit
 │
 └─ BIND MOUNT: the working tree is shared host <-> container
```

Key properties:

- **Fresh stateless iteration.** Each loop iteration is a brand-new `opencode run` with
  no carried context. The builder re-orients every time from durable state: `git log`,
  `git status`, `TASK.md`, the project's normatives doc, and the named plan. **The git
  history IS the durable progress** — robust to crashes, restarts, and recreation.

- **The working tree is bind-mounted.** Builder and orchestrator edit the *same files*.
  Coordination rule:
  - **Sequence on shared files.** Don't have the orchestrator edit a file the running
    loop is also touching. Pause/relaunch the loop, or carve disjoint work.
  - **Parallelize only on DISJOINT trees.** Two workers are safe only when their file
    sets don't overlap.
  - **The orchestrator owns ALL git.** The builder never commits, pushes, or opens PRs.
    It only edits files and emits a stop token. Branching, committing, PRs, and merges
    are exclusively the orchestrator's job.

- **Secrets arrive read-only at `/run/secrets`, never baked into the image.** They live
  only in the gitignored `secrets/` dir on the host and are mounted in at runtime.

---

## 3. The four artifacts the orchestrator maintains

### (a) `TASK.md` — the per-unit brief

The single most important artifact. Gitignored scratch (it is *not* part of the repo's
durable history). The orchestrator **rewrites it before every unit**. The builder treats
it as the spec and re-reads it every fresh iteration. A good `TASK.md` is **self-contained**:

- The unit name and a one-line goal.
- A pointer to the authoritative plan/section (so the builder reads design, not memory).
- **The exact files to create** — and, when cloning, the existing file to use as the
  template/pattern.
- **In scope / out of scope** — name the files that must NOT be touched (other packages,
  `go.mod`, etc.). Cheap models wander; fence them.
- **Acceptance checks** — concrete, verifiable conditions for "done."
- **The exact gate commands**, including the **pinned linter version** (run the CI-pinned
  linter, not the host's — see §9).
- A "do NOT commit/push/PR" reminder (the orchestrator owns git).

See [`templates/TASK.example.md`](./templates/TASK.example.md) for a worked unit.

### (b) `BUILD_PROMPT.md` — the fixed builder prompt

The *same* prompt every iteration; per-unit detail lives in `TASK.md`, which the prompt
tells the builder to read. Its structure (see [`templates/BUILD_PROMPT.md`](./templates/BUILD_PROMPT.md)):

- **ORIENT** — read `TASK.md`, the normatives doc (`CLAUDE.md`), the named plan; run
  `git log`/`git status`; assume nothing carried over (stateless).
- **SCOPE** — do ONE coherent unit exactly as `TASK.md` defines; if acceptance checks are
  already satisfied, emit `[goal:complete]` and stop.
- **BUILD** — follow the framework's own tooling/skills; never reverse-engineer an API
  from memory. Includes the critical **tool-use discipline** (see §7 / §9 — small edits,
  never giant single writes that truncate silently).
- **GATE** — run every gate (build/test/lint/...); green or it is NOT done.
- **REPORT** — do NOT commit/push/PR. Emit `[goal:complete]` when all gates pass and
  acceptance holds; emit `[goal:blocked]` + a one-line reason with `file:line` evidence
  if stuck. A failing/skipped gate is never "done."

### (c) The model config (`opencode.json`)

Provider config with **the model pinned explicitly** (so plugins/subagents can't silently
route to a paid model), `permission: allow` (safe inside the container, avoids unattended
stalls), and `instructions` pointing at the project's normatives doc. This is also where
you add an OpenAI-compatible provider for a new model (see §5).

### (d) The verification & git discipline the orchestrator follows

Not a file — your operating procedure (§4, §7). The builder produces; you verify, fix the
hard parts, and merge.

### The signaling contract

The loop watches the builder's output for two tokens on their own line:

- **`[goal:complete]`** — the builder believes the unit's acceptance checks pass and all
  gates are green. The loop stops. **This is an untrusted claim** — your verification
  begins here.
- **`[goal:blocked]`** — the builder hit something it can't resolve; it appends a
  one-line reason with `file:line` evidence. The loop stops; you investigate.

The loop also pattern-matches rate-limit/quota signals to drive primary→fallback failover
and backoff. Anything else just ends the iteration and a fresh one starts.

---

## 4. The orchestrator cadence

One trip around this loop per unit. Nothing advances on the builder's say-so — only on
your verified gate.

1. **Polish the plan.** Read the design source of truth + the plan; settle scope.
   Decompose into **one coherent unit per builder iteration** (one module, one surface,
   one endpoint-set). For a novel/cross-cutting slice, **build the first reference unit
   yourself** so the builder has a correct pattern to clone.

2. **Write a small, self-contained `TASK.md` unit** (§3a). Pin the gates. Fence the scope.

3. **Launch the loop:** `./.devcontainer/run.sh`. (Re-running recreates the container.)

4. **Monitor periodically (~every 20-30 min). Don't hover.** Each check:
   - `status.txt` — terminal state? (`COMPLETE`/`BLOCKED` → stop, start verifying.)
   - `run.log` — is the iteration counter advancing? **Note:** `run.log` only flushes
     **between** iterations, so mid-iteration it looks frozen. To see progress *during* an
     iteration, watch **git/file movement** (`git status`, file mtimes), not the log.
   - `git log`/`git status` — are files actually changing, or "active but unproductive"?
   - Confirm the log still shows the **model you pinned** (no silent reroute).
   - **Stall test:** if `run.log` mtime *and* git are frozen for many minutes *and* the
     counter isn't advancing, the iteration is wedged (compaction stall / hung process).
   - **Rule: leave a moving loop alone; relaunch a stale one** (`./.devcontainer/run.sh`).
     Over-monitoring tempts premature kills that cost more than they save. Each recreate
     restarts the model's orientation from scratch (~minutes wasted on a slow model).

5. **On `[goal:complete]`, INDEPENDENTLY VERIFY — don't trust the signal** (§7):
   - Run **all gates** yourself (build/test/lint/validate), per module, in the real build
     mode.
   - **Read the actual code** — handler/component bodies — for the stub trap (registered
     surface vs. code that calls the real backend; orphaned dead-code helpers). Count
     artifacts vs. claims.
   - **Run it live** where applicable (HTTP smoke, a browser via Playwright, the
     framework's inspector against fixtures). "It compiles/renders" ≠ "it works."
   - **Fix trivial defects yourself.** A one-line correctness fix is cheaper than ten
     cheap iterations. Feed larger evidenced findings back as the next narrow `TASK.md`.

6. **Commit, PR, merge.** Branch off `main`, commit with a message stating what the
   builder produced **and** what you fixed on review, open a PR, get CI green,
   squash-merge. Never push straight to `main`. The merged history is the durable record.

7. **Next unit.** Edit `TASK.md`, re-run `run.sh`.

**Division of labor that works:** the builder is good at **clones** —
storage layers, handlers, CLI commands, pure functions, anything with an existing
pattern to copy. It is weaker on **novel/security/UI** work and anything requiring
cross-cutting judgment — keep those with the orchestrator, or hand the builder a correct
reference unit first.

---

## 5. Model strategy (TIME-SENSITIVE — read this)

> **There is a ~4-hour window of FREE access to GLM-5.2 (SOTA) via the Hugging Face
> Router → Fireworks provider. While that window is open, GLM-5.2 is the PRIMARY builder
> model.** It is OpenAI-compatible, has a **1M-token context** (configure ~500k to stay
> safe), and is dramatically more reliable than the fallbacks below.

| Role | Model | Provider | baseURL | Model id | Notes |
|---|---|---|---|---|---|
| **Primary (free window)** | **GLM-5.2** | HF Router → Fireworks | `https://router.huggingface.co/v1` | `zai-org/GLM-5.2:fireworks-ai` | Auth via `HF_TOKEN`. 1M ctx → set `limit.context` ~500000. **Use while the free window lasts.** |
| Fallback | NVIDIA NIM Nemotron-3 Ultra 550B | NVIDIA NIM | `https://integrate.api.nvidia.com/v1` | `nvidia/nemotron-3-ultra-550b-a55b` | Auth via `NVIDIA_API_KEY`. **UNRELIABLE** for this loop (EngineCore errors + malformed `<function=...>` tool-call syntax). Prefer GLM-5.2 while available. |
| Candidate 2nd fallback | "Nemotron Nano 2.5" / "nimo 2.5" | NVIDIA NIM | `https://integrate.api.nvidia.com/v1` | (smaller Nemotron variant) | Try if the big Nemotron keeps failing. |

After the free GLM-5.2 window expires, fall back to NVIDIA NIM Nemotron — but expect the
reliability problems noted above.

### Adding an OpenAI-compatible model to `opencode.json`

The provider block shape (key + `baseURL` + `apiKey` env ref + a `models` map carrying the
context/output `limit`):

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "model": "huggingface/zai-org/GLM-5.2:fireworks-ai",   // pinned PRIMARY
  "provider": {
    "huggingface": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "HuggingFace Router",
      "options": {
        "baseURL": "https://router.huggingface.co/v1",
        "apiKey": "{env:HF_TOKEN}"
      },
      "models": {
        "zai-org/GLM-5.2:fireworks-ai": {
          "name": "GLM-5.2 (Fireworks)",
          "limit": { "context": 500000, "output": 65536 }
        }
      }
    },
    "nvidia": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "NVIDIA NIM",
      "options": {
        "baseURL": "https://integrate.api.nvidia.com/v1",
        "apiKey": "{env:NVIDIA_API_KEY}"
      },
      "models": {
        "nemotron-3-ultra-550b-a55b": {
          "name": "Nemotron-3 Ultra 550B",
          "limit": { "context": 131072, "output": 65536 }
        }
      }
    }
  },
  "instructions": ["CLAUDE.md"],
  "permission": { "bash": "allow", "write": "allow", "edit": "allow", "read": "allow" }
}
```

Notes:
- Set `limit.output` generously (32k+). Reasoning tokens eat the output budget; too small
  a budget is a major cause of silent write truncation (§9).
- **Pin the model explicitly** (`"model": "<provider>/<id>"`) so a plugin/subagent can't
  silently reroute to a paid model. Confirm the pinned model in `run.log` while monitoring.

### How the token reaches the model

1. **`run.sh`** (host) writes the secret from an env var into the gitignored `secrets/`
   dir, e.g.:
   ```bash
   if [ -n "${HF_TOKEN:-}" ]; then printf '%s' "$HF_TOKEN" > "$DC/secrets/hf_token"; fi
   chmod 600 "$DC/secrets/hf_token"
   ```
2. **`entrypoint.sh`** (container) reads it from `/run/secrets/<name>` and exports it for
   both `exec "$@"` and the login-shell loop (`/etc/profile.d`):
   ```bash
   if [ -f "$SEC/hf_token" ]; then
     HF_TOKEN="$(tr -d '[:space:]' < "$SEC/hf_token")"; export HF_TOKEN
     echo "export HF_TOKEN=$HF_TOKEN" > /etc/profile.d/hf.sh; chmod 644 /etc/profile.d/hf.sh
   fi
   ```
3. **`opencode.json`** references it as `"apiKey": "{env:HF_TOKEN}"`.

> **NEVER** put a real token value in any committed file. Use placeholders (`hf_xxx`,
> `{env:HF_TOKEN}`). Tokens live ONLY in the gitignored `secrets/` dir. The full git
> range is secret-scanned (gitleaks) — never commit a token-shaped fixture.

---

## 6. The container artifacts

| File | Role |
|---|---|
| `Dockerfile` | Toolchain image only (language runtime, Node if there's a UI, the CI-pinned linter, `git`/`gh`, Bun, opencode). **No secrets.** |
| `devcontainer.json` | Bind mounts: the repo, the read-only `secrets/`, the runtime `var/`. |
| `entrypoint.sh` | Runtime secret injection (opencode `auth.json`, model API key, gh token), git identity, `OPENCODE_CONFIG` pointer, warm the build cache. |
| `run.sh` | **Host** script: prep secrets → build image → recreate container → launch loop detached. |
| `loop.sh` | PID 1: fresh `opencode run --agent build "<BUILD_PROMPT>"` per iteration; primary/fallback model; rate-limit backoff; stop-token detection; writes `status.txt` + `run.log`. |
| `opencode.json` | Provider config + pinned model + `permission: allow` + `instructions`. |
| `BUILD_PROMPT.md` | The fixed per-iteration builder prompt. |
| `TASK.md` | The per-unit brief (gitignored). |
| `secrets/`, `var/` | Gitignored — injected secrets + run logs/status. Never committed. |

Sanitized copies of all of these are in [`templates/`](./templates/).

**`run.sh` essentials** (see template):
- Prep secrets into `secrets/` (gh token via `gh auth token`, opencode `auth.json`,
  model API keys from env), `chmod 600` them.
- `docker build` the image.
- `docker rm -f` the old container, then `docker run -d` a new one with the task env
  (`COMMAND`, `MODEL`, `MAX_ITERS`) and the three bind mounts. The loop is the container's
  main process; `--restart on-failure:N` self-heals crashes while a clean
  `[goal:complete]` exits 0 and stays stopped.
- **Never `docker start` a stopped container** — that resurrects the old loop with stale
  task env and causes dual-loop API contention (§9).

**`loop.sh` essentials** (see template):
- `export PATH=...` at the top (login shells reset it → "command not found").
- `export OPENCODE_CONFIG=...` pointing at the repo's `opencode.json` (it's under
  `.devcontainer/`, not the repo root, so opencode won't auto-load it from cwd).
- Per iteration: pick primary/fallback model, run `opencode run --model "$M" --agent build
  "$BUILD_PROMPT"`, append output to `run.log`, grep for the stop tokens and rate-limit
  signals, sleep briefly, repeat until `MAX_ITERS`.
- Always `exit 0` on a clean end so the restart policy doesn't relaunch after completion.

---

## 7. Verification — never trust the completion signal

This is the heart of the method. **Low-tier models lie about completion** — confidently.
"All done" with empty UIs; "all servers built" with stub handlers returning hardcoded
literals; "all gates pass" when only the root module was checked. `[goal:complete]` ends
one iteration; it is not a fact.

Verify on three axes, all orchestrator-owned:

**Machine gates.** Build, test (with `-race` if available), lint, validate — green or not
done. Two traps: **verify per module** in the real build mode (a root build can miss
sub-modules), and **count artifacts vs. claims** ("200+ tests" was 21 files).

**Read the actual code (the stub trap — the big one).** A cheap builder generates handler
shells that return empty/literal structs and never call the real backend, while writing
the *real* integration as **orphaned dead-code helpers**. Detect it:
- Compare # registered surfaces vs. # handlers that actually call the real client/store.
- Grep handlers for returns of empty/literal structs with no client call.
- Check the registration function even *receives* its backend dependency — a
  `Register(srv)` with no deps passed in is a guaranteed stub.
- "Calls the client" ≠ "correct": a second pass checks endpoints/params/shape against a
  behavioral oracle (a reference implementation), not just "an API call exists."

**Live validation.** It isn't done until it renders/responds. Run the real surface — HTTP
smoke checks, a browser via Playwright, the framework's inspector against fixtures — and
exercise each tool/UI. "It compiles and is themed" hides functionally-dead surfaces.

**Multiple polish passes, in order** (one pass is never enough; cheap output converges
over repetition): correctness → usability/product surface → docs (godoc) → README. Drive
each as its own narrow loop or orchestrator review. **Write findings down** with
`file:line` evidence and feed them back as targeted `TASK.md` units.

> When the bar must be highest, escalate to **capable-model multi-agent workflows**:
> foundation barrier → one agent per unit → an adversarial critic per unit → a final
> integration gate you re-verify. Reserve it for the hard/novel slices and the final
> quality pass — not for cheap breadth.

---

## 8. Adapting to a new project (`go-slides-mcp` checklist)

1. **Copy the templates** into the new repo's `.devcontainer/` and adjust:
   - In `run.sh`: set `REPO=`, `IMAGE=`, `NAME=` to the new project; the bind-mount
     target path (`/workspace/<proj>`); which secrets to prep.
   - In `devcontainer.json`: the `workspaceFolder` / mount target.
   - In `loop.sh` / `entrypoint.sh`: the `cd` path and `OPENCODE_CONFIG` path.
2. **Set the build/test/lint gate commands** for the new project (in `BUILD_PROMPT.md`'s
   GATE step and in each `TASK.md`). Pin the linter version to whatever CI uses.
3. **Point `instructions` at the project's normatives doc** (e.g. `CLAUDE.md`) in
   `opencode.json`, and make the project actually have one (see "doc-first" below).
4. **Set the primary model** — GLM-5.2 while the free window lasts (§5), else the
   Nemotron fallback. Write the token into `secrets/` via `run.sh` + env.
5. **Write the first `TASK.md`** — one small, self-contained, clonable unit. If the
   project has no reference pattern yet, build that first unit yourself.
6. **Doc-first prerequisite.** Before delegating a line, the project should have: a
   binding normatives doc (product invariants + repo layout + conventions + a
   forbidden-practices list), a phased plan with machine-checkable acceptance criteria,
   and a stated priority chain (which artifact wins on conflict). The weaker the builder,
   the more upfront spec it needs — the docs do the reasoning the model can't.

### How `TASK.md` briefs reference binding normatives

The builder is low-context and re-reads everything each iteration, so `TASK.md` must point
it at the binding rules rather than relying on training recall. Two patterns worth
calling out (from the reference project's `CLAUDE.md`):

- **Phase/implementor contract (ref §4.2).** Every new surface (HTTP endpoint, RPC method,
  CLI command) comes with its test/smoke check *in the same unit*; a phase is "done" only
  when its acceptance criteria pass AND the smoke/gate scripts are green with FAIL=0. Put
  the exact gate commands and the "add a smoke check for any new endpoint" rule into
  `BUILD_PROMPT.md` so the builder can't skip them.
- **Extensibility seams (ref §4.4).** Any subsystem with plausible alternate backends goes
  behind an interface + factory + registry, with one driver per subdirectory, drivers
  self-registering and pulled in via blank import at the binary entry point. When a unit
  builds such a subsystem, `TASK.md` names the reference implementation to clone (e.g.
  "follow `internal/storage/`") so the builder replicates the established seam rather than
  inventing a single concrete type.

The rule: `TASK.md` *names the normative and the file to clone*; the builder follows the
pattern instead of guessing.

---

## 9. Gotchas (learned the hard way)

- **Dual-loop contention.** Switch tasks by **recreating** via `run.sh`
  (`docker rm -f` + `docker run`), **never** `docker start` a stopped container + exec. A
  stopped container's PID 1 is the old loop with the *original* task env; restarting it
  runs the finished task, and a second loop on top means two loops hammering the same
  rate-limited API. Detect: `ps -eo pid,etime,cmd | grep -E "loop.sh|opencode run"`
  showing >1 agent process.

- **Don't run host `make build` while the loop runs.** A host build can leave a
  **host-arch binary** in `./bin/` that the Linux container can't exec (or vice versa),
  breaking the container's gates. Let the container own its build artifacts; don't race it
  from the host.

- **Pinned-linter version mismatch.** Run the **CI-pinned** linter, not the host's. The
  reference project pins `golangci-lint` and installs it *with the project's Go toolchain*
  (the linter refuses to run when built with an older Go than the `go.mod` target). In
  `TASK.md` use the pinned install, e.g.:
  ```bash
  GOBIN=/tmp/gcl go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
  /tmp/gcl/golangci-lint run ./...   # run WITHOUT | tail (you want to see all findings)
  ```

- **Large single `write` calls truncate silently.** Reasoning models over an
  OpenAI-compatible API emit `write` tool-calls whose JSON is **cut off mid-string**, so a
  big whole-file write **fails silently** ("Invalid input for tool write: JSON parsing
  failed / Expected '}'" / "Unterminated string") while small edits succeed. It looks
  "active but unproductive" — steps advance, no files change. **Fix:** set `limit.output`
  high (32k+) AND mandate in the prompt: grow files with small targeted `edit` calls
  (skeleton first, then append one function at a time, each payload < ~150 lines); split
  big files into several small ones. Detect by grepping `run.log` for `Unterminated
  string` / `Invalid Tool`.

- **Compaction wedge.** A long single session grows until it auto-compacts; if the
  compaction call itself gets rate-limited, the session is wedged (context full → can't
  continue; compaction throttled → can't recover) and never *exits*, so the fresh-stateless
  outer loop can't restart it. Symptom: the agent process is alive but file mtimes and
  `run.log` are frozen for many minutes. **Fix:** keep per-iteration scope small (one
  unit) so a session never grows big enough to compact; relaunch (recreate) the wedged
  loop. Stall detection is your job.

- **`run.log` only flushes between iterations.** Mid-iteration it looks frozen even when
  work is happening — judge progress by git/file movement, not the log, during an
  iteration.

- **Prompt-change timing.** Editing `BUILD_PROMPT.md` or `TASK.md` only takes effect on
  the **next fresh iteration**, never the one already running.

- **Rate-limit failover from the log, not stdout.** Agents often log the rate limit
  internally; a killed process emits nothing. A loop that only switches models on a stdout
  match keeps retrying a dead primary. Detect from the log file, or pin the working model
  as primary.

- **Toolchain env sabotage.** A stray build-flag env var can make every build fail and
  *look* like the builder produced broken code. When a gate fails, check the environment
  before blaming the builder. Keep the toolchain env minimal/neutral.

- **Check `git status` for stray binaries before `git add`.** A build can leave a binary
  (`./bin/...`) in the tree. Prefer **explicit `git add <files>`** over `git add -A`, and
  scan the diff. Combined with the full-range secret scan, **never commit a token-shaped
  fixture** — gitleaks scans history, not just the tip.

- **Stale render in screenshots.** Use a unique port per unit; reusing a port serves a
  cached/stale surface. Clean-build before validating (build caches serve stale bundles).

- **Test rot from signature drift.** When handlers change shape, old tests break and block
  the merge gate — budget a test-fix pass.

---

## See also

- The Claude Code skill **`orchestrate-autonomous-build`** — the canonical, full write-up
  of this method (container lifecycle, stateless-loop rationale, the verify-don't-trust
  depth, the multi-agent quality wave, and the complete gotchas checklist).
- [`templates/`](./templates/) — sanitized, project-agnostic copies of every config file.
```
