package exportstore

import "path/filepath"

const (
	// exportsDirName is the deterministic export directory under the workspace.
	exportsDirName = "exports"
	// PPTXMIMEType is the MIME type for PowerPoint .pptx files.
	PPTXMIMEType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
)

// ExportPath returns the deterministic workspace path for one deck export.
func ExportPath(workspace, deckID string) string {
	return filepath.Join(workspace, exportsDirName, deckID+".pptx")
}

// DeckResourceURI returns the deck:// resource URI for one exported deck.
func DeckResourceURI(deckID string) string {
	return "deck://export/" + deckID + ".pptx"
}
