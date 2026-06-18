package contracts

import "encoding/json"

// CardSection is a top-level card accepting grids, two-columns, or nested
// cards. Body must be non-empty (validation, later unit). Mirror of
// scene.CardSection. Children nest recursively.
type CardSection struct {
	// Header is the section title.
	Header string `json:"header,omitempty"`
	// Body is the section body children (must be non-empty).
	Body []SlideNode `json:"body,omitempty"`
}

func (CardSection) slideNodeKind() Kind { return KindCardSection }

// MarshalJSON injects the "card_section" kind; Body marshals through each
// child's own MarshalJSON (kind injected per child).
func (c *CardSection) MarshalJSON() ([]byte, error) {
	return marshalNode(KindCardSection, *c)
}

// UnmarshalJSON dispatches Body through UnmarshalSlideNode so the section
// nests recursively (CONVENTIONS §3).
func (c *CardSection) UnmarshalJSON(data []byte) error {
	type raw struct {
		Header string            `json:"header,omitempty"`
		Body   []json.RawMessage `json:"body,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	c.Header = r.Header
	body, err := unmarshalNodes(r.Body)
	if err != nil {
		return err
	}
	c.Body = body
	return nil
}

func init() {
	registerNodeKind(KindCardSection, func() SlideNode { return &CardSection{} })
}
