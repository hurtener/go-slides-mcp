// Command go-slides-mcp serves the Deckard MCP protocol over stdio or HTTP.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/hurtener/dockyard/runtime/apps"
	"github.com/hurtener/dockyard/runtime/server"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/comment"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
	"github.com/hurtener/go-slides-mcp/internal/exportstore"
	"github.com/hurtener/go-slides-mcp/internal/handlers"
	"github.com/hurtener/go-slides-mcp/internal/recipe"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// uiBundles embeds the built single-file Svelte surfaces. `dockyard build` runs
// Vite first so each dist/index.html exists at compile time. The all: prefix is
// required so the embed includes the bundle (RFC §14).
//
//go:embed all:web/apps/deck-preview/dist
//go:embed all:web/apps/deck-overview/dist
//go:embed all:web/apps/slide-editor/dist
var uiBundles embed.FS

// The three UI surfaces. URIs are honored verbatim (CLAUDE.md §7).
const (
	deckPreviewName  = "deck-preview"
	deckPreviewURI   = "ui://go-slides-mcp/deck-preview/index.html"
	deckOverviewName = "deck-overview"
	deckOverviewURI  = "ui://go-slides-mcp/deck-overview/index.html"
	slideEditorName  = "slide-editor"
	slideEditorURI   = "ui://go-slides-mcp/slide-editor/index.html"
)

// httpAddr is the address the HTTP transport listens on when
// DOCKYARD_TRANSPORT=http. DOCKYARD_HTTP_ADDR overrides it.
const httpAddr = "127.0.0.1:8080"

var buildInfo = contracts.BuildInfo{Name: "go-slides-mcp", Version: "0.1.0"}

func main() {
	// A text slog handler — readable local logs (Dockyard convention).
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Serve until the process is interrupted (Ctrl-C) or the host closes the
	// transport.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	srv, err := server.New(server.Info{
		Name:    buildInfo.Name,
		Title:   "Go Slides Mcp",
		Version: buildInfo.Version,
	}, &server.Options{Logger: logger})
	if err != nil {
		logger.Error("create server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	workspace, err := workspaceDir()
	if err != nil {
		logger.Error("resolve workspace", slog.String("error", err.Error()))
		os.Exit(1)
	}
	deps := handlers.ToolDeps{
		Store:     deck.NewMemoryStore(),
		Souls:     soul.NewMemoryRegistry(),
		Assets:    asset.NewMemoryStore(),
		Comments:  comment.NewMemoryStore(),
		Recipes:   recipe.NewMemoryStore(),
		Session:   &handlers.SessionState{},
		BuildInfo: buildInfo,
		Workspace: workspace,
		Brand:     loadBrand(logger),
		Logger:    logger,
	}

	if err := exportstore.RegisterResources(srv, deps.Workspace); err != nil {
		logger.Error("register resources", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Apps register BEFORE tools so each tool's .UI(name) resolves (CLAUDE.md §7).
	if err := registerApps(srv); err != nil {
		logger.Error("register apps", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := handlers.RegisterTools(srv, deps); err != nil {
		logger.Error("register tools", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := serve(ctx, srv, logger); err != nil {
		logger.Error("serve", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// registerApps installs the embedded single-file Svelte surfaces as MCP Apps.
// Each surface's tools attach via .UI(name); the deny-by-default CSP just works
// because the bundles are single-file (no external origins).
func registerApps(srv *server.Server) error {
	type uiApp struct{ name, uri, title, path string }
	for _, a := range []uiApp{
		{deckPreviewName, deckPreviewURI, "Deckard — Deck preview", "web/apps/deck-preview/dist/index.html"},
		{deckOverviewName, deckOverviewURI, "Deckard — Deck overview", "web/apps/deck-overview/dist/index.html"},
		{slideEditorName, slideEditorURI, "Deckard — Slide editor", "web/apps/slide-editor/dist/index.html"},
	} {
		html, err := fs.ReadFile(uiBundles, a.path)
		if err != nil {
			return err
		}
		if err := apps.Register(srv, apps.App{URI: a.uri, Name: a.name, Title: a.title, HTML: html}); err != nil {
			return err
		}
	}
	return nil
}

// loadBrand resolves the white-label brand config at startup. DECKARD_BRAND_TOKENS
// points at a JSON file ({title, defaultTheme, tokens, allowThemeSwitch}); unset
// or unreadable falls back to the built-in Deckard White brand (a warning, never
// a failure — a bad brand file must not take the server down).
func loadBrand(logger *slog.Logger) contracts.AppBrand {
	def := contracts.AppBrand{Title: "Deckard Slides", DefaultTheme: "deckard-white", AllowThemeSwitch: true}
	path := os.Getenv("DECKARD_BRAND_TOKENS")
	if path == "" {
		return def
	}
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Warn("brand tokens unreadable; using Deckard White", slog.String("path", path), slog.String("error", err.Error()))
		return def
	}
	// Parse with a *bool for allowThemeSwitch so an omitted key means "shown"
	// (the contract field is a plain bool for clean codegen).
	var raw struct {
		Title            string            `json:"title"`
		DefaultTheme     string            `json:"defaultTheme"`
		Tokens           map[string]string `json:"tokens"`
		AllowThemeSwitch *bool             `json:"allowThemeSwitch"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		logger.Warn("brand tokens invalid JSON; using Deckard White", slog.String("path", path), slog.String("error", err.Error()))
		return def
	}
	b := contracts.AppBrand{
		Title:            raw.Title,
		DefaultTheme:     raw.DefaultTheme,
		Tokens:           raw.Tokens,
		AllowThemeSwitch: raw.AllowThemeSwitch == nil || *raw.AllowThemeSwitch,
	}
	if b.Title == "" {
		b.Title = def.Title
	}
	if b.DefaultTheme == "" {
		b.DefaultTheme = def.DefaultTheme
	}
	logger.Info("loaded white-label brand", slog.String("path", path), slog.String("title", b.Title), slog.String("theme", b.DefaultTheme))
	return b
}

func workspaceDir() (string, error) {
	if dir := os.Getenv("DECKARD_WORKSPACE"); dir != "" {
		return dir, nil
	}
	return os.Getwd()
}

// serve brings up the transport named by DOCKYARD_TRANSPORT. An unset or
// "stdio" value serves stdio; "http" serves the streamable-HTTP transport. An
// unrecognised value is a clean, explained failure rather than a silent
// fallback.
func serve(ctx context.Context, srv *server.Server, logger *slog.Logger) error {
	switch transport := os.Getenv("DOCKYARD_TRANSPORT"); transport {
	case "", "stdio":
		return srv.ServeStdio(ctx)
	case "http":
		return serveHTTP(ctx, srv, logger)
	default:
		return errors.New("unsupported DOCKYARD_TRANSPORT " + transport + " (want \"stdio\" or \"http\")")
	}
}

// serveHTTP serves the streamable-HTTP transport. The HTTP security posture is
// the runtime's secure default — DNS-rebinding and cross-origin protection both
// on (runtime/server.DefaultHTTPSecurity). The listen address is httpAddr,
// overridable with DOCKYARD_HTTP_ADDR.
func serveHTTP(ctx context.Context, srv *server.Server, logger *slog.Logger) error {
	handler, err := srv.HTTPHandler(nil)
	if err != nil {
		return err
	}
	addr := httpAddr
	if override := os.Getenv("DOCKYARD_HTTP_ADDR"); override != "" {
		addr = override
	}
	httpSrv := &http.Server{Addr: addr, Handler: handler}
	go func() {
		<-ctx.Done()
		_ = httpSrv.Close()
	}()
	logger.Info("serving streamable-HTTP transport", slog.String("addr", addr))
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
