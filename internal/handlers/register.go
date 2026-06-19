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
	if err := tool.New[contracts.AddSlideInput, contracts.AddSlideOutput]("add_slide").
		Describe("Add one slide to a deck and return its validation result.").
		Handler(h.addSlide).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.UpdateSlideInput, contracts.UpdateSlideOutput]("update_slide").
		Describe("Replace one slide in a deck and return its validation result.").
		Handler(h.updateSlide).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.GetSlideInput, contracts.GetSlideOutput]("get_slide").
		Describe("Get one slide by deck and slide ID and return its validation result.").
		Handler(h.getSlide).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.RemoveSlideInput, contracts.RemoveSlideOutput]("remove_slide").
		Describe("Remove one slide from a deck.").
		Handler(h.removeSlide).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.ReorderSlidesInput, contracts.ReorderSlidesOutput]("reorder_slides").
		Describe("Replace a deck's slide order.").
		Handler(h.reorderSlides).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.DuplicateSlideInput, contracts.DuplicateSlideOutput]("duplicate_slide").
		Describe("Duplicate one slide in a deck.").
		Handler(h.duplicateSlide).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.EditSlideNodeInput, contracts.EditSlideNodeOutput]("edit_slide_node").
		Describe("Replace one slide node at a structural path.").
		Handler(h.editSlideNode).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.InsertSlideNodeInput, contracts.InsertSlideNodeOutput]("insert_slide_node").
		Describe("Insert one slide node at a structural path.").
		Handler(h.insertSlideNode).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.RemoveSlideNodeInput, contracts.RemoveSlideNodeOutput]("remove_slide_node").
		Describe("Remove one slide node at a structural path.").
		Handler(h.removeSlideNode).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.DuplicateSlideNodeInput, contracts.DuplicateSlideNodeOutput]("duplicate_slide_node").
		Describe("Duplicate one slide node at a structural path.").
		Handler(h.duplicateSlideNode).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.MoveSlideNodeInput, contracts.MoveSlideNodeOutput]("move_slide_node").
		Describe("Move one slide node between structural paths.").
		Handler(h.moveSlideNode).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.BootstrapSoulInput, contracts.BootstrapSoulOutput]("bootstrap_soul").
		Describe("Bootstrap one design soul from a small set of token overrides.").
		Handler(h.bootstrapSoul).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.RefineSoulInput, contracts.RefineSoulOutput]("refine_soul").
		Describe("Refine one stored soul with targeted token overrides.").
		Handler(h.refineSoul).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.ListSoulsInput, contracts.ListSoulsOutput]("list_souls").
		Describe("List every stored soul summary.").
		Handler(h.listSouls).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.GetSoulInput, contracts.GetSoulOutput]("get_soul").
		Describe("Get one stored soul, optionally including its style guide.").
		Handler(h.getSoul).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.GetDesignTokensInput, contracts.GetDesignTokensOutput]("get_design_tokens").
		Describe("Get the flattened design token list for one stored soul.").
		Handler(h.getDesignTokens).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.UploadAssetInput, contracts.UploadAssetOutput]("upload_asset").
		Describe("Upload one binary asset and return its stored metadata.").
		Handler(h.uploadAsset).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.ListAssetsInput, contracts.ListAssetsOutput]("list_assets").
		Describe("List every stored asset metadata summary.").
		Handler(h.listAssets).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.GetAssetInput, contracts.GetAssetOutput]("get_asset").
		Describe("Get one stored asset metadata summary by asset ID.").
		Handler(h.getAsset).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.DeleteAssetInput, contracts.DeleteAssetOutput]("delete_asset").
		Describe("Delete one stored asset by asset ID.").
		Handler(h.deleteAsset).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.ValidateSlideIRInput, contracts.ValidateSlideIROutput]("validate_slide_ir").
		Describe("Validate one slide IR snapshot without storage.").
		Handler(h.validateSlideIR).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.ValidateSlideInput, contracts.ValidateSlideOutput]("validate_slide").
		Describe("Validate one stored slide by deck and slide ID.").
		Handler(h.validateSlide).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.ValidateDeckForExportInput, contracts.ValidateDeckForExportOutput]("validate_deck_for_export").
		Describe("Validate every slide in one deck and report export blockers.").
		Handler(h.validateDeckForExport).
		Register(srv); err != nil {
		return err
	}
	if err := tool.New[contracts.GetSessionInput, contracts.GetSessionOutput]("get_session").
		Describe("Get the active in-memory workspace session and build info.").
		Handler(h.getSession).
		Register(srv); err != nil {
		return err
	}

	return nil
}
