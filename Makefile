# Deckard (go-slides-mcp) — canonical build / test / lint / dev commands.
# Single Go module, single binary, three UI surfaces. GOFLAGS MUST stay empty
# (never -mod=mod — it breaks workspace mode silently). See CLAUDE.md §4/§11.

BINARY := bin/go-slides-mcp
SURFACES := deck-preview deck-overview slide-editor

.PHONY: generate
generate: ## regenerate JSON Schema + TS from Go contracts (idempotent)
	GOFLAGS="" dockyard generate

.PHONY: validate
validate: ## fast quality gate — fails on stale contracts / blockers
	GOFLAGS="" dockyard validate

.PHONY: build
build: ## CGo-free static binary with the UI embedded
	GOFLAGS="" CGO_ENABLED=0 go build -o $(BINARY) .

.PHONY: vet
vet: ## go vet ./...
	GOFLAGS="" go vet ./...

.PHONY: test
test: ## go test -race ./...
	GOFLAGS="" go test -race ./...

.PHONY: dockyard-test
dockyard-test: ## full contract + spec + capability gate
	GOFLAGS="" dockyard test

.PHONY: lint
lint: ## pinned golangci-lint (v2.12.2)
	golangci-lint run

.PHONY: web
web: ## type-check + build every surface + the design system
	cd web/design-system && npx svelte-check --tsconfig ./tsconfig.json && npx vite build
	@for s in $(SURFACES); do \
		echo "=== web: $$s ==="; \
		(cd web/apps/$$s && npx svelte-check --tsconfig ./tsconfig.json && npx vite build); \
	done

.PHONY: web-check
web-check: ## type-check every surface + the design system (no build)
	cd web/design-system && npx svelte-check --tsconfig ./tsconfig.json
	@for s in $(SURFACES); do (cd web/apps/$$s && npx svelte-check --tsconfig ./tsconfig.json); done

.PHONY: check-mirror
check-mirror: ## AGENTS.md == CLAUDE.md
	@diff -q AGENTS.md CLAUDE.md && echo "Mirror OK" || (echo "ERROR: AGENTS.md != CLAUDE.md" && exit 1)

.PHONY: fmt
fmt: ## gofmt -s -w everything (excludes _ref)
	gofmt -s -w $$(git ls-files '*.go')

.PHONY: fmt-check
fmt-check: ## fail if anything is unformatted
	@out=$$(gofmt -l $$(git ls-files '*.go')); [ -z "$$out" ] || (echo "unformatted:"; echo "$$out"; exit 1)

.PHONY: preflight
preflight: generate validate build test ## the CI gate (contract-clean + build + race tests)

.PHONY: gate
gate: fmt-check generate validate vet build test dockyard-test check-mirror ## full local gate (green or NOT done)
	@git diff --exit-code internal/contracts >/dev/null || (echo "ERROR: stale codegen in internal/contracts" && exit 1)
	@echo "GATE GREEN"

.PHONY: dev
dev: ## watch + regenerate + rebuild + restart + auto inspector
	dockyard dev

.PHONY: help
help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'
