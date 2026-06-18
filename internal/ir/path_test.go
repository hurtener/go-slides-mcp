package ir

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func heroT(title string) *contracts.Hero { return &contracts.Hero{Title: title} }

// nested builds: nodes[0]=two_column{left:[hero A], right:[card{body:[hero B, hero C]}]}
func nested() *contracts.Slide {
	return &contracts.Slide{ID: "s", Nodes: []contracts.SlideNode{
		&contracts.TwoColumn{
			Left:  []contracts.SlideNode{heroT("A")},
			Right: []contracts.SlideNode{&contracts.Card{Header: "card", Body: []contracts.SlideNode{heroT("B"), heroT("C")}}},
		},
	}}
}

func titleAt(t *testing.T, s *contracts.Slide, p Path) string {
	t.Helper()
	n, err := Resolve(s, p)
	if err != nil {
		t.Fatalf("resolve %v: %v", p, err)
	}
	h, ok := n.(*contracts.Hero)
	if !ok {
		t.Fatalf("resolve %v: want *Hero, got %T", p, n)
	}
	return h.Title
}

func TestResolveNested(t *testing.T) {
	s := nested()
	if got := titleAt(t, s, Path{"nodes", 0, "left", 0}); got != "A" {
		t.Fatalf("left[0]=%q want A", got)
	}
	if got := titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 1}); got != "C" {
		t.Fatalf("right card body[1]=%q want C", got)
	}
	// float64 indices (JSON shape) resolve too.
	if got := titleAt(t, s, Path{"nodes", float64(0), "right", float64(0), "body", float64(0)}); got != "B" {
		t.Fatalf("float-index body[0]=%q want B", got)
	}
}

func TestSetInsertRemove(t *testing.T) {
	s := nested()
	if err := Set(s, Path{"nodes", 0, "left", 0}, heroT("A2")); err != nil {
		t.Fatal(err)
	}
	if got := titleAt(t, s, Path{"nodes", 0, "left", 0}); got != "A2" {
		t.Fatalf("after set, left[0]=%q want A2", got)
	}

	// Insert at body[1] (between B and C); append at end too.
	if err := Insert(s, Path{"nodes", 0, "right", 0, "body", 1}, heroT("X")); err != nil {
		t.Fatal(err)
	}
	if err := Insert(s, Path{"nodes", 0, "right", 0, "body", 3}, heroT("Z")); err != nil { // append at len
		t.Fatal(err)
	}
	got := []string{
		titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 0}),
		titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 1}),
		titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 2}),
		titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 3}),
	}
	want := []string{"B", "X", "C", "Z"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("after inserts body=%v want %v", got, want)
		}
	}

	// Remove the inserted X (now at body[1]).
	rem, err := Remove(s, Path{"nodes", 0, "right", 0, "body", 1})
	if err != nil {
		t.Fatal(err)
	}
	if rem.(*contracts.Hero).Title != "X" {
		t.Fatalf("removed %q want X", rem.(*contracts.Hero).Title)
	}
	if got := titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 1}); got != "C" {
		t.Fatalf("after remove body[1]=%q want C", got)
	}
}

func TestDuplicateIsDeepCopy(t *testing.T) {
	s := nested()
	dup, err := Duplicate(s, Path{"nodes", 0, "right", 0, "body", 0}) // duplicate B
	if err != nil {
		t.Fatal(err)
	}
	// The duplicate lands at body[1]; original B stays at body[0], C shifts to body[2].
	if got := titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 1}); got != "B" {
		t.Fatalf("dup at body[1]=%q want B", got)
	}
	if got := titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 2}); got != "C" {
		t.Fatalf("C shifted to body[2], got %q", got)
	}
	// Mutating the clone must not touch the original (deep copy).
	dup.(*contracts.Hero).Title = "B-clone"
	if got := titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 0}); got != "B" {
		t.Fatalf("original B mutated to %q — not a deep copy", got)
	}
}

func TestMoveSameSliceForwardAndBackward(t *testing.T) {
	// body = [B, C, D]; move B (idx0) to idx2 → [C, D, B] (destination after source).
	s := &contracts.Slide{Nodes: []contracts.SlideNode{
		&contracts.Card{Body: []contracts.SlideNode{heroT("B"), heroT("C"), heroT("D")}},
	}}
	if err := Move(s, Path{"nodes", 0, "body", 0}, Path{"nodes", 0, "body", 2}); err != nil {
		t.Fatal(err)
	}
	order := []string{
		titleAt(t, s, Path{"nodes", 0, "body", 0}),
		titleAt(t, s, Path{"nodes", 0, "body", 1}),
		titleAt(t, s, Path{"nodes", 0, "body", 2}),
	}
	if order[0] != "C" || order[1] != "D" || order[2] != "B" {
		t.Fatalf("forward move got %v want [C D B]", order)
	}
	// Move B (now idx2) back to idx0 → [B, C, D].
	if err := Move(s, Path{"nodes", 0, "body", 2}, Path{"nodes", 0, "body", 0}); err != nil {
		t.Fatal(err)
	}
	if titleAt(t, s, Path{"nodes", 0, "body", 0}) != "B" {
		t.Fatalf("backward move failed: %v", []string{
			titleAt(t, s, Path{"nodes", 0, "body", 0}),
			titleAt(t, s, Path{"nodes", 0, "body", 1}),
			titleAt(t, s, Path{"nodes", 0, "body", 2}),
		})
	}
}

func TestMoveAcrossContainers(t *testing.T) {
	s := nested() // left:[A], right card body:[B,C]
	// Move A from left[0] into the card body at index 0.
	if err := Move(s, Path{"nodes", 0, "left", 0}, Path{"nodes", 0, "right", 0, "body", 0}); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(s, Path{"nodes", 0, "left", 0}); err == nil {
		t.Fatal("left should now be empty")
	}
	if got := titleAt(t, s, Path{"nodes", 0, "right", 0, "body", 0}); got != "A" {
		t.Fatalf("moved A should be body[0], got %q", got)
	}
}

func TestPathErrors(t *testing.T) {
	s := nested()
	cases := []struct {
		name string
		p    Path
	}{
		{"empty", Path{}},
		{"odd", Path{"nodes"}},
		{"not-nodes", Path{"slides", 0}},
		{"oob-index", Path{"nodes", 9}},
		{"bad-leg", Path{"nodes", 0, "middle", 0}},
		{"leaf-has-no-child", Path{"nodes", 0, "left", 0, "body", 0}}, // hero has no "body"
		{"bad-index-type", Path{"nodes", "two"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Resolve(s, c.p); err == nil {
				t.Fatalf("path %v should error", c.p)
			}
		})
	}
}
