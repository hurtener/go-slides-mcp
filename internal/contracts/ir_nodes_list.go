package contracts

import (
	"encoding/json"
	"fmt"
)

// ListKind selects a list style (mirrors pptx-go's scene.ListKind).
type ListKind string

// List styles (wire values per CONVENTIONS §2 example).
const (
	ListBullet    ListKind = "bullet"
	ListNumber    ListKind = "number"
	ListChecklist ListKind = "checklist"
)

// ListItem is one entry in a List. Mirror of scene.ListItem.
type ListItem struct {
	// Text is the item's content.
	Text RichText `json:"text,omitempty"`
	// Level is the nesting depth (0 = top-level).
	Level int `json:"level,omitempty"`
	// Checked marks a checklist item as done (ListChecklist only).
	Checked bool `json:"checked,omitempty"`
}

// List is a bullet, numbered, or checklist list. Mirror of scene.List. The
// JSON field for the list style is "listKind" (not "kind", which is reserved
// for the node discriminator — CONVENTIONS §2).
type List struct {
	// Kind is the list style.
	Kind ListKind `json:"listKind,omitempty"`
	// Items is the ordered list entries.
	Items []ListItem `json:"items,omitempty"`
}

func (List) slideNodeKind() Kind { return KindList }

// MarshalJSON injects the "list" kind discriminator via marshalNode.
func (l *List) MarshalJSON() ([]byte, error) { return marshalNode(KindList, *l) }

// UnmarshalJSON strict-decodes a List and each ListItem so unknown keys are a
// hard error naming the offending field(s). The injected "kind" discriminator
// is explicitly allowed.
func (l *List) UnmarshalJSON(data []byte) error {
	type listWire struct {
		Kind  ListKind          `json:"listKind,omitempty"`
		Items []json.RawMessage `json:"items,omitempty"`
	}
	var wire listWire
	if err := strictUnmarshal(data, &wire, "kind"); err != nil {
		return err
	}
	l.Kind = wire.Kind
	if wire.Items != nil {
		l.Items = make([]ListItem, len(wire.Items))
		for i, raw := range wire.Items {
			if err := strictUnmarshal(raw, &l.Items[i]); err != nil {
				return fmt.Errorf("items[%d]: %w", i, err)
			}
		}
	}
	return nil
}

func init() { registerNodeKind(KindList, func() SlideNode { return &List{} }) }
