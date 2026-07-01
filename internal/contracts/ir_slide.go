package contracts

import "encoding/json"

// Slide is one slide: a stable ID, a structural layout, an ordered node tree,
// and optional speaker notes. Nodes and Notes route through the codec
// (CONVENTIONS §2/§3): marshaling injects each node's "kind" via the node's
// own MarshalJSON; unmarshaling dispatches each node through
// UnmarshalSlideNode (recursion) and each run through TextRun's unmarshaler.
type Slide struct {
	// ID is the slide identifier (stable, agent/human addressable).
	ID string `json:"id"`
	// Layout is the structural intent, mapping to a master layout.
	Layout LayoutKind `json:"layout,omitempty"`
	// Align sets how the slide's body content sits in the frame: vertical
	// top|center|bottom|justify, horizontal left|center|right. Empty = top-left.
	// Per-node align fields override the horizontal axis for individual blocks.
	Align Alignment `json:"align,omitempty"`
	// Variant selects a theme variant for this slide: "light" (default) or
	// "dark". VariantDark renders a dark canvas with light text — the engine
	// derives a legible dark palette from the active soul automatically.
	// Omitting the field is identical to "light".
	Variant Variant `json:"variant,omitempty"`
	// Background is an optional full-bleed fill drawn behind all content.
	// Set kind to "color" (solid soul color role), "gradient" (two-stop
	// linear gradient of soul color roles), or "asset" (full-bleed picture).
	// Nil (the default) draws nothing — byte-identical to pre-existing slides.
	Background *Background `json:"background,omitempty"`
	// Archetype is the slide's role in the deck (cover/section/content/dark/
	// closing). It selects the soul's per-archetype background+decoration
	// recipe (R13.12) when the slide sets no explicit Background/decorations.
	// Empty = inferred from Layout/Variant/position by the composer.
	Archetype SlideArchetype `json:"archetype,omitempty"`
	// Section is the per-slide chrome eyebrow label (R3, opt-in). When the
	// deck chrome is enabled, this string appears in the top section-eyebrow
	// band above the body on this slide (e.g. "01 — Direction"). Empty means
	// no eyebrow is drawn on this slide even when deck chrome is enabled.
	Section string `json:"section,omitempty"`
	// Nodes is the slide's top-level node tree.
	Nodes []SlideNode `json:"nodes,omitempty" jsonschema:"ordered list of slide nodes. Each node is a JSON object with a kind discriminator, one of: hero|heading|prose|list|callout|quote|chip|table|two_column|grid|card|card_section|flow|chart|code_block|image|divider|arrow|section_divider|decoration. Every RichText field (heading.text, prose.paragraphs[], quote.text, callout.body, list items[].text, table headers/rows, flow steps[].label/detail) is an ARRAY of FLAT runs — [{\"text\":\"hi\"}] or [{\"text\":\"38% lower\",\"bold\":true,\"italic\":true,\"color\":{\"token\":\"accent\"}}]; there is NO nested style object and keys are lowercase. Variant keys are NOT named kind: list uses listKind (bullet|ordered|checklist) + items[].text; callout uses calloutKind (info|tip|warning|success|error) + title + body. flow uses steps[] of {label:RichText, detail:RichText, icon?} — NOT title/body. Examples: heading {\"kind\":\"heading\",\"level\":2,\"text\":[{\"text\":\"Highlights\"}]}; list {\"kind\":\"list\",\"listKind\":\"bullet\",\"items\":[{\"text\":[{\"text\":\"first\"}]}]}; callout {\"kind\":\"callout\",\"calloutKind\":\"tip\",\"title\":\"Heads up\",\"body\":[{\"text\":\"detail\"}]}; flow {\"kind\":\"flow\",\"steps\":[{\"label\":[{\"text\":\"Start\"}],\"detail\":[{\"text\":\"kick off\"}]}]}. Call describe_node for the full per-kind shape."`
	// Notes is the speaker notes as RichText — a JSON ARRAY of FLAT runs, e.g.
	// [{"text":"speak to "},{"text":"this point","bold":true}]. There is no
	// nested "style" object and every key is lowercase.
	Notes RichText `json:"notes,omitempty" jsonschema:"speaker notes as RichText: a JSON array of FLAT runs — [{\"text\":\"plain\"}] or [{\"text\":\"emphasis\",\"bold\":true,\"italic\":true,\"color\":{\"token\":\"accent\"}}]. There is NO nested style object and keys are lowercase (text, typeRole, bold, italic, underline, strike, code, link, href, color)."`
	// Footnotes are slide-level source/citation/disclaimer lines (R14.12),
	// each a RichText paragraph, rendered into a reserved band at the bottom
	// of the slide (muted role; the body region shrinks to reserve the
	// band). Lines past the engine's region cap are dropped and warned.
	// Empty/nil (the default) draws no band — byte-identical to slides
	// authored before R14.12.
	Footnotes []RichText `json:"footnotes,omitempty" jsonschema:"slide-level source/citation/disclaimer lines: a JSON array of RichText paragraphs, each itself an array of FLAT runs — [[{\"text\":\"Source: internal telemetry, 2026.\"}]]. Rendered into a reserved band at the bottom of the slide; there is NO nested style object and keys are lowercase."`
}

// MarshalJSON routes Nodes through each child's MarshalJSON and Notes through
// TextRun's MarshalJSON. The plain helper type has no MarshalJSON, so this
// never recurses.
func (s *Slide) MarshalJSON() ([]byte, error) {
	type plain struct {
		ID         string         `json:"id"`
		Layout     LayoutKind     `json:"layout,omitempty"`
		Align      Alignment      `json:"align,omitempty"`
		Variant    Variant        `json:"variant,omitempty"`
		Background *Background    `json:"background,omitempty"`
		Archetype  SlideArchetype `json:"archetype,omitempty"`
		Section    string         `json:"section,omitempty"`
		Nodes      []SlideNode    `json:"nodes,omitempty"`
		Notes      RichText       `json:"notes,omitempty"`
		Footnotes  []RichText     `json:"footnotes,omitempty"`
	}
	return json.Marshal(plain{
		ID:         s.ID,
		Layout:     s.Layout,
		Align:      s.Align,
		Variant:    s.Variant,
		Background: s.Background,
		Archetype:  s.Archetype,
		Section:    s.Section,
		Nodes:      s.Nodes,
		Notes:      s.Notes,
		Footnotes:  s.Footnotes,
	})
}

// UnmarshalJSON dispatches Nodes through UnmarshalSlideNode (recursive) and
// Notes through TextRun's unmarshaler.
func (s *Slide) UnmarshalJSON(data []byte) error {
	type raw struct {
		ID         string            `json:"id"`
		Layout     LayoutKind        `json:"layout,omitempty"`
		Align      Alignment         `json:"align,omitempty"`
		Variant    Variant           `json:"variant,omitempty"`
		Background *Background       `json:"background,omitempty"`
		Archetype  SlideArchetype    `json:"archetype,omitempty"`
		Section    string            `json:"section,omitempty"`
		Nodes      []json.RawMessage `json:"nodes,omitempty"`
		Notes      RichText          `json:"notes,omitempty"`
		Footnotes  []RichText        `json:"footnotes,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	s.ID = r.ID
	s.Layout = r.Layout
	s.Align = r.Align
	s.Variant = r.Variant
	s.Background = r.Background
	s.Archetype = r.Archetype
	s.Section = r.Section
	nodes, err := unmarshalNodes(r.Nodes)
	if err != nil {
		return err
	}
	s.Nodes = nodes
	s.Notes = r.Notes
	s.Footnotes = r.Footnotes
	return nil
}

// SlideDoc is the deck-of-slides wrapper: a title, deck-level chrome, and an
// ordered slide list. Slides route through Slide's own marshal/unmarshal, so
// node trees and notes are codec-handled throughout.
type SlideDoc struct {
	// Title is the deck title.
	Title string `json:"title,omitempty"`
	// Chrome is the deck-level slide chrome configuration (R3). When
	// Chrome.Enabled is true the engine draws section-eyebrow and footer bands.
	// The zero value (Enabled == false) draws no chrome — byte-identical to
	// decks authored before R3.
	Chrome DeckChrome `json:"chrome,omitempty"`
	// Slides is the deck's slides, in order.
	Slides []Slide `json:"slides,omitempty"`
}

// MarshalJSON routes Slides through each Slide's MarshalJSON. The plain
// helper type has no MarshalJSON, so this never recurses.
func (d *SlideDoc) MarshalJSON() ([]byte, error) {
	type plain struct {
		Title  string     `json:"title,omitempty"`
		Chrome DeckChrome `json:"chrome,omitempty"`
		Slides []Slide    `json:"slides,omitempty"`
	}
	return json.Marshal(plain{Title: d.Title, Chrome: d.Chrome, Slides: d.Slides})
}

// UnmarshalJSON routes Slides through each Slide's UnmarshalJSON (which in
// turn dispatches nodes recursively).
func (d *SlideDoc) UnmarshalJSON(data []byte) error {
	type raw struct {
		Title  string     `json:"title,omitempty"`
		Chrome DeckChrome `json:"chrome,omitempty"`
		Slides []Slide    `json:"slides,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	d.Title = r.Title
	d.Chrome = r.Chrome
	d.Slides = r.Slides
	return nil
}
