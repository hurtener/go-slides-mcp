package handlers

import (
	"context"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func TestGetSessionReturnsBuildInfoAndEmptyState(t *testing.T) {
	h := testHandlers()
	got, err := h.getSession(context.Background(), contracts.GetSessionInput{})
	if err != nil {
		t.Fatalf("getSession: %v", err)
	}
	if got.Structured.BuildInfo.Name != "go-slides-mcp" || got.Structured.BuildInfo.Version != "test" {
		t.Fatalf("getSession build info = %+v", got.Structured.BuildInfo)
	}
	if got.Structured.ActiveDeckID != "" || got.Structured.ActiveSoulID != "" {
		t.Fatalf("getSession active state = %+v", got.Structured)
	}
	if len(got.Structured.OpenPanels) != 0 {
		t.Fatalf("getSession open panels = %v, want empty", got.Structured.OpenPanels)
	}
}
