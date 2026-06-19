package handlers

import (
	"context"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func (h *handlers) getSession(_ context.Context, _ contracts.GetSessionInput) (tool.Result[contracts.GetSessionOutput], error) {
	activeDeckID, activeSoulID, openPanels := h.deps.Session.Snapshot()
	out := contracts.GetSessionOutput{
		ActiveDeckID: activeDeckID,
		ActiveSoulID: activeSoulID,
		OpenPanels:   openPanels,
		BuildInfo:    h.deps.BuildInfo,
	}
	return tool.Result[contracts.GetSessionOutput]{Text: fmt.Sprintf("Loaded session for %s %s.", out.BuildInfo.Name, out.BuildInfo.Version), Structured: out}, nil
}
