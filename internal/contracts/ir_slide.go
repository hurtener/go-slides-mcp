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
	// Nodes is the slide's top-level node tree.
	Nodes []SlideNode `json:"nodes,omitempty"`
	// Notes is the speaker notes.
	Notes RichText `json:"notes,omitempty"`
}

// MarshalJSON routes Nodes through each child's MarshalJSON and Notes through
// TextRun's MarshalJSON. The plain helper type has no MarshalJSON, so this
// never recurses.
func (s *Slide) MarshalJSON() ([]byte, error) {
	type plain struct {
		ID     string      `json:"id"`
		Layout LayoutKind  `json:"layout,omitempty"`
		Nodes  []SlideNode `json:"nodes,omitempty"`
		Notes  RichText    `json:"notes,omitempty"`
	}
	return json.Marshal(plain{
		ID:     s.ID,
		Layout: s.Layout,
		Nodes:  s.Nodes,
		Notes:  s.Notes,
	})
}

// UnmarshalJSON dispatches Nodes through UnmarshalSlideNode (recursive) and
// Notes through TextRun's unmarshaler.
func (s *Slide) UnmarshalJSON(data []byte) error {
	type raw struct {
		ID     string            `json:"id"`
		Layout LayoutKind        `json:"layout,omitempty"`
		Nodes  []json.RawMessage `json:"nodes,omitempty"`
		Notes  RichText          `json:"notes,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	s.ID = r.ID
	s.Layout = r.Layout
	nodes, err := unmarshalNodes(r.Nodes)
	if err != nil {
		return err
	}
	s.Nodes = nodes
	s.Notes = r.Notes
	return nil
}

// SlideDoc is the deck-of-slides wrapper: a title plus an ordered slide list.
// Slides route through Slide's own marshal/unmarshal, so node trees and notes
// are codec-handled throughout.
type SlideDoc struct {
	// Title is the deck title.
	Title string `json:"title,omitempty"`
	// Slides is the deck's slides, in order.
	Slides []Slide `json:"slides,omitempty"`
}

// MarshalJSON routes Slides through each Slide's MarshalJSON. The plain
// helper type has no MarshalJSON, so this never recurses.
func (d *SlideDoc) MarshalJSON() ([]byte, error) {
	type plain struct {
		Title  string  `json:"title,omitempty"`
		Slides []Slide `json:"slides,omitempty"`
	}
	return json.Marshal(plain{Title: d.Title, Slides: d.Slides})
}

// UnmarshalJSON routes Slides through each Slide's UnmarshalJSON (which in
// turn dispatches nodes recursively).
func (d *SlideDoc) UnmarshalJSON(data []byte) error {
	type raw struct {
		Title  string  `json:"title,omitempty"`
		Slides []Slide `json:"slides,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	d.Title = r.Title
	d.Slides = r.Slides
	return nil
}
