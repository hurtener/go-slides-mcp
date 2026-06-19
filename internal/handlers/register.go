package handlers

import (
	"github.com/hurtener/dockyard/runtime/server"
	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// RegisterTools declares and registers every tool this server exposes.
func RegisterTools(srv *server.Server, deps ToolDeps) error {
	h := &handlers{deps: deps}

	if err := tool.New[contracts.CreateDeckInput, contracts.CreateDeckOutput]("create_deck").
		Describe("Create a new deck and return its deck summary.").
		Handler(h.createDeck).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.ListDecksInput, contracts.ListDecksOutput]("list_decks").
		Describe("List every stored deck summary.").
		Handler(h.listDecks).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.GetDeckInput, contracts.GetDeckOutput]("get_deck").
		Describe("Get one deck by ID or slug and return its current summary.").
		Handler(h.getDeck).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.DeleteDeckInput, contracts.DeleteDeckOutput]("delete_deck").
		Describe("Delete one deck by ID or slug.").
		Handler(h.deleteDeck).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.SetDeckChromeInput, contracts.SetDeckChromeOutput]("set_deck_chrome").
		Describe("Replace a deck's header and footer chrome.").
		Handler(h.setDeckChrome).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.SetDeckSectionsInput, contracts.SetDeckSectionsOutput]("set_deck_sections").
		Describe("Replace a deck's named section grouping.").
		Handler(h.setDeckSections).
		Register(srv); err != nil {
		return err
	}

	return nil
}
