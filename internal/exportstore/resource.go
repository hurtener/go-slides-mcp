package exportstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hurtener/dockyard/runtime/server"
)

// ErrNotFound reports a missing exported deck file.
var ErrNotFound = errors.New("exportstore: not found")

// RegisterResources registers the deck:// export resource family.
func RegisterResources(srv *server.Server, workspace string) error {
	return srv.AddResourceTemplate(server.ResourceTemplateDef{
		URITemplate: "deck://export/{id}.pptx",
		Name:        "deck-export",
		Title:       "Deckard export (.pptx)",
		Description: "The exported PowerPoint for a deck, by deck id.",
		MIMEType:    PPTXMIMEType,
	}, func(ctx context.Context, uri string) (server.ResourceContent, error) {
		_ = ctx
		id, err := ParseDeckID(uri)
		if err != nil {
			return server.ResourceContent{}, err
		}
		buf, err := os.ReadFile(ExportPath(workspace, id))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return server.ResourceContent{}, fmt.Errorf("%s: %w", uri, ErrNotFound)
			}
			return server.ResourceContent{}, err
		}
		return server.ResourceContent{MIMEType: PPTXMIMEType, Blob: buf}, nil
	})
}

// ParseDeckID maps a concrete deck:// URI back to its deck identifier.
func ParseDeckID(uri string) (string, error) {
	const prefix = "deck://export/"
	const suffix = ".pptx"
	if !strings.HasPrefix(uri, prefix) || !strings.HasSuffix(uri, suffix) {
		return "", fmt.Errorf("invalid deck resource uri %q", uri)
	}
	name := strings.TrimSuffix(strings.TrimPrefix(uri, prefix), suffix)
	if name == "" || strings.ContainsRune(name, filepath.Separator) {
		return "", fmt.Errorf("invalid deck resource uri %q", uri)
	}
	return name, nil
}
