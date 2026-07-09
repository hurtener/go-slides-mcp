package conformance

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/render"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/go-slides-mcp/internal/validate"
)

// TestConformanceCorpus is R14.19's standing acceptance: every archetype
// fixture, rendered through every soul variant, must satisfy INV-1..INV-5.
// A subtest name is "<archetype>/<soul>" so a CI failure pinpoints both the
// regressing archetype class and the soul variant that broke it.
func TestConformanceCorpus(t *testing.T) {
	souls, err := Souls()
	if err != nil {
		t.Fatalf("Souls() error = %v", err)
	}

	for _, sv := range souls {
		for _, fx := range Archetypes {
			t.Run(fx.Name+"/"+sv.Name, func(t *testing.T) {
				t.Parallel()
				checkArchetype(t, fx, sv)
			})
		}
	}
}

// checkArchetype renders fx through sv and asserts INV-1..INV-5.
func checkArchetype(t *testing.T, fx Fixture, sv SoulVariant) {
	t.Helper()
	doc := fx.Build()

	// INV-1: render-ok + valid PPTX.
	buf, stats, err := render.Render(doc, sv.Soul)
	if err != nil {
		t.Fatalf("INV-1: Render() error = %v", err)
	}
	assertValidPPTX(t, buf)

	// INV-2: no overflow / safe-area warnings.
	if len(stats.Warnings) != 0 {
		t.Errorf("INV-2: stats.Warnings = %v, want empty", stats.Warnings)
	}
	if len(stats.LayoutWarnings) != 0 {
		t.Errorf("INV-2: stats.LayoutWarnings = %#v, want empty", stats.LayoutWarnings)
	}

	// INV-3: contrast — no error-severity issue, per slide or at the soul's
	// theme level.
	for _, sc := range stats.Colors {
		for _, iss := range validate.AuditSlideColors(sc) {
			if iss.Severity == validate.SeverityError {
				t.Errorf("INV-3: AuditSlideColors(%s) error: %s", sc.SlideID, iss.Message)
			}
		}
	}
	for _, iss := range validate.AuditTheme(sv.Soul.Theme) {
		if iss.Severity == validate.SeverityError {
			t.Errorf("INV-3: AuditTheme(%s) error: %s", sv.Name, iss.Message)
		}
	}

	// INV-4: soul fidelity — resolved per-slide colors equal the soul's
	// declared tokens.
	mismatches, err := render.SoulColorFidelity(doc, sv.Soul)
	if err != nil {
		t.Fatalf("INV-4: SoulColorFidelity() error = %v", err)
	}
	if len(mismatches) != 0 {
		t.Errorf("INV-4: SoulColorFidelity mismatches = %#v, want none", mismatches)
	}

	// INV-5: byte-identical re-render. The worker-count axis is covered by
	// render's own package-internal TestRenderDeterministicAcrossWorkerCounts
	// (renderWithWorkers is unexported); this package stays on the public
	// render.Render entry point per the R14.19 spec.
	second, _, err := render.Render(doc, sv.Soul)
	if err != nil {
		t.Fatalf("INV-5: second Render() error = %v", err)
	}
	if !bytes.Equal(buf, second) {
		t.Error("INV-5: re-render bytes differ")
	}
}

// TestConformanceCorpus_DecorPolicyIsApplied guards the whole corpus against a
// silent failure mode: if the R13-D decor policy ever became a no-op, every
// INV-1..5 check above would still pass (the invariants hold on a flat deck),
// so the corpus would give false "on bar" assurance. Anchor it: the SAME
// content fixture rendered through a bootstrapped (decorated) soul must emit
// strictly MORE shapes than through the flat built-in Deckard White — the paper
// fill + full-bleed dot texture the policy injects are real, counted shapes.
func TestConformanceCorpus_DecorPolicyIsApplied(t *testing.T) {
	doc := contentDoc()

	_, flat, err := render.Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("flat Render() error = %v", err)
	}
	decorated, err := soul.Bootstrap(soul.BootstrapParams{Name: "Decor Guard", Accent: "3B5BDB"})
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	_, dec, err := render.Render(doc, decorated)
	if err != nil {
		t.Fatalf("decorated Render() error = %v", err)
	}
	if dec.Shapes <= flat.Shapes {
		t.Errorf("decorated render Shapes = %d, want > flat Shapes = %d (decor policy not applied?)", dec.Shapes, flat.Shapes)
	}
}

// assertValidPPTX is a local copy of internal/render's assertValidPPTX
// helper (unexported, can't be called across packages): the rendered bytes
// must be a well-formed zip carrying [Content_Types].xml.
func assertValidPPTX(t *testing.T, buf []byte) {
	t.Helper()
	if len(buf) == 0 {
		t.Fatal("Render() returned empty bytes")
	}
	r, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}
	for _, f := range r.File {
		if f.Name == "[Content_Types].xml" {
			return
		}
	}
	t.Fatal("rendered zip missing [Content_Types].xml")
}

// TestConformanceCorpus_IsExtensible documents (and guards) R14.19's
// "one-fixture addition" clause: adding a new archetype is exactly one
// Fixture{Name, Build} entry appended to the Archetypes slice in corpus.go —
// no changes to this test file, the soul builders, or the invariant runner.
// This test is a tripwire against silent registry shrinkage/duplication, not
// a growth mechanism itself.
func TestConformanceCorpus_IsExtensible(t *testing.T) {
	seen := make(map[string]bool, len(Archetypes))
	for _, fx := range Archetypes {
		if fx.Name == "" {
			t.Fatal("Archetypes entry has an empty Name")
		}
		if fx.Build == nil {
			t.Fatalf("Archetypes[%q].Build is nil", fx.Name)
		}
		if seen[fx.Name] {
			t.Fatalf("Archetypes has a duplicate Name %q", fx.Name)
		}
		seen[fx.Name] = true
	}
	if len(Archetypes) == 0 {
		t.Fatal("Archetypes is empty")
	}
}

func TestSerifSoulContentFitFixtures(t *testing.T) {
	serif, err := soul.Bootstrap(soul.BootstrapParams{
		Name:        "Conformance Serif",
		HeadingFont: "Playfair Display",
		BodyFont:    "Lora",
	})
	if err != nil {
		t.Fatalf("Bootstrap serif soul error = %v", err)
	}
	for _, fx := range []struct {
		name  string
		build func() contracts.SlideDoc
	}{
		{name: "chip-row", build: chipRowDoc},
		{name: "pricing-offer-card", build: pricingOfferCardDoc},
	} {
		t.Run(fx.name, func(t *testing.T) {
			doc := fx.build()
			_, stats, err := render.Render(doc, serif)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if len(stats.Warnings) != 0 || len(stats.LayoutWarnings) != 0 {
				t.Fatalf("warnings = %v / %#v, want none", stats.Warnings, stats.LayoutWarnings)
			}
		})
	}
}
