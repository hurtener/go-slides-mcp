// Command go-slides-mcp serves the Deckard MCP protocol over stdio or HTTP.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/hurtener/dockyard/runtime/server"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/comment"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
	"github.com/hurtener/go-slides-mcp/internal/exportstore"
	"github.com/hurtener/go-slides-mcp/internal/handlers"
	"github.com/hurtener/go-slides-mcp/internal/soul"
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
		Session:   &handlers.SessionState{},
		BuildInfo: buildInfo,
		Workspace: workspace,
		Logger:    logger,
	}

	if err := exportstore.RegisterResources(srv, deps.Workspace); err != nil {
		logger.Error("register resources", slog.String("error", err.Error()))
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
