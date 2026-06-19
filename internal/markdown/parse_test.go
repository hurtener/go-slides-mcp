package markdown

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func TestParseHeadings(t *testing.T) {
	nodes, _ := Parse("# Title\n## Subtitle")
	if len(nodes) != 2 {
		t.Fatalf("got %d nodes, want 2", len(nodes))
	}
	h1, ok := nodes[0].(*contracts.Heading)
	if !ok || h1.Level != 1 {
		t.Fatalf("node[0] = %#v, want Heading level 1", nodes[0])
	}
	h2, ok := nodes[1].(*contracts.Heading)
	if !ok || h2.Level != 2 {
		t.Fatalf("node[1] = %#v, want Heading level 2", nodes[1])
	}
}

func TestParseGroupsConsecutiveBullets(t *testing.T) {
	nodes, _ := Parse("- one\n- two\n- three")
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1 List", len(nodes))
	}
	list, ok := nodes[0].(*contracts.List)
	if !ok || list.Kind != contracts.ListBullet {
		t.Fatalf("node[0] = %#v, want bullet List", nodes[0])
	}
	if len(list.Items) != 3 {
		t.Fatalf("got %d items, want 3", len(list.Items))
	}
}

func TestParseNumberedList(t *testing.T) {
	nodes, _ := Parse("1. first\n2. second")
	list, ok := nodes[0].(*contracts.List)
	if !ok || list.Kind != contracts.ListNumber || len(list.Items) != 2 {
		t.Fatalf("got %#v, want numbered List with 2 items", nodes[0])
	}
}

func TestParseQuoteAndProse(t *testing.T) {
	nodes, _ := Parse("> a wise quote\n\nA plain paragraph line\nsecond line of it")
	if _, ok := nodes[0].(*contracts.Quote); !ok {
		t.Fatalf("node[0] = %#v, want Quote", nodes[0])
	}
	prose, ok := nodes[1].(*contracts.Prose)
	if !ok || len(prose.Paragraphs) != 1 {
		t.Fatalf("node[1] = %#v, want Prose with 1 paragraph", nodes[1])
	}
}

func TestParseMixedSequence(t *testing.T) {
	nodes, _ := Parse("# H\n\n- a\n- b\n\ntext here")
	if len(nodes) != 3 {
		t.Fatalf("got %d nodes, want 3 (heading, list, prose)", len(nodes))
	}
}

// Regression: "1." (ordered marker with no content after the dot) must not panic.
func TestParseNumberItemNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Parse panicked on '1.': %v", r)
		}
	}()
	nodes, _ := Parse("1.")
	if len(nodes) == 0 {
		t.Fatal("expected the line to fold into a node")
	}
}
