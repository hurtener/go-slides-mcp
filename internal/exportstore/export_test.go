package exportstore

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
)

func TestExportWritesDeterministicValidPPTX(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	doc := testDoc()

	firstPath, firstStats, err := Export(workspace, "deck_export", doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Export() first error = %v", err)
	}
	secondPath, secondStats, err := Export(workspace, "deck_export", doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Export() second error = %v", err)
	}
	if firstPath != secondPath {
		t.Fatalf("Export() path = %q, want stable %q", secondPath, firstPath)
	}
	if firstStats.Slides == 0 || secondStats.Slides == 0 {
		t.Fatalf("Export() slides stats = %d/%d, want > 0", firstStats.Slides, secondStats.Slides)
	}

	buf, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", firstPath, err)
	}
	if len(buf) == 0 {
		t.Fatal("exported file is empty")
	}
	if _, err := pptx.NewFromBytes(buf); err != nil {
		t.Fatalf("pptx.NewFromBytes() error = %v", err)
	}

	expectedPath, err := filepath.Abs(ExportPath(workspace, "deck_export"))
	if err != nil {
		t.Fatalf("filepath.Abs() error = %v", err)
	}
	if firstPath != expectedPath {
		t.Fatalf("Export() path = %q, want %q", firstPath, expectedPath)
	}

	again, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) second error = %v", secondPath, err)
	}
	if !bytes.Equal(buf, again) {
		t.Fatal("exported bytes differ across repeated exports")
	}
}

func TestParseDeckID(t *testing.T) {
	t.Parallel()

	got, err := ParseDeckID("deck://export/deck_123.pptx")
	if err != nil {
		t.Fatalf("ParseDeckID() error = %v", err)
	}
	if got != "deck_123" {
		t.Fatalf("ParseDeckID() = %q, want deck_123", got)
	}
	if _, err := ParseDeckID("deck://export/.pptx"); err == nil {
		t.Fatal("ParseDeckID() empty id error = nil, want non-nil")
	}
}

func testDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Export Test",
		Slides: []contracts.Slide{{
			ID:     "slide-1",
			Layout: contracts.LayoutTitleContent,
			Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Agenda"}}},
				&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Body"}}}},
			},
		}},
	}
}
