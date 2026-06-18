# Build target — current unit (EXAMPLE — copy to .devcontainer/TASK.md and rewrite per unit)

> Self-contained brief for a fresh, stateless builder session. Do exactly this unit,
> gate it green, stop. Do NOT commit/push/PR — the orchestrator owns git.
> SCOPE: name the surface and FENCE it — list exactly what NOT to touch.

## Unit <ID>: <one-line goal>

**Plan:** `docs/plans/<plan-file>.md` — section <N>. This unit builds ONLY <scope>.
The <other parts> are SEPARATE units the orchestrator owns — do NOT build them.

State any invariants the builder must hold (determinism, no global state, interface seam,
naming, etc.). Name the reference implementation to clone if there is an established pattern.

---

### In scope — create these files

**1. `path/to/file_one.ext`**
- What it must contain (types, functions, signatures). Quote the exact signatures so the
  builder doesn't invent them.
- The exact semantics to implement (the tests below assert these).

**2. `path/to/file_one_test.ext`**
- Exhaustive table tests against FIXED inputs and expected outputs. Spell out the cases.

---

### Out of scope (DO NOT TOUCH)
- List every package/file/dir the builder must not modify (other modules, go.mod, the UI,
  cmd/, etc.). Cheap models wander — fence them explicitly.

### Acceptance checks
- Concrete, verifiable conditions for "done" (e.g. "function X returns Y for input Z";
  "every exported symbol has a doc comment"; "smoke check N is green").

### Gates (all must pass from repo root)
```
# ADAPT to your project. PIN the linter to the CI version — do not use the host's.
gofmt -l path/to/dir            # must print NOTHING
go vet ./path/...
go test -race ./path/...
GOBIN=/tmp/gcl go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
/tmp/gcl/golangci-lint run ./path/... --timeout=5m   # run WITHOUT | tail — see all findings
go build ./...
```
When all gates are green AND the acceptance checks hold, output exactly `[goal:complete]`
on its own line and STOP. Do NOT commit/push/PR.
