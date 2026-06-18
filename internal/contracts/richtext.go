package contracts

import "encoding/json"

// RichText is an ordered list of styled runs — the inline-text content type
// used across the slide IR (CONVENTIONS §4). It marshals as a JSON array of
// TextRun objects; each run is a flat { text, typeRole?, bold?, ... , color? }.
type RichText []TextRun

// TextRun is one styled run of text. Its JSON shape is flat (CONVENTIONS §4):
// the Style fields are inlined into the run object and Color is omitted when
// unset (meaning the token "primary").
type TextRun struct {
	// Text is the literal run content.
	Text string
	// Style carries typography and inline formatting.
	Style RunStyle
	// Color is the run color; the zero value means the token "primary".
	Color TextColor
}

// RunStyle mirrors pptx-go's scene.RunStyle: typography role plus inline
// formatting flags and a link target.
type RunStyle struct {
	// TypeRole selects a typography scale role (TypeBody, TypeH2, ...).
	TypeRole TypeRole
	// Bold toggles bold weight.
	Bold bool
	// Italic toggles italic style.
	Italic bool
	// Underline toggles underline.
	Underline bool
	// Strike toggles strikethrough.
	Strike bool
	// Code marks the run as inline code (mono + tint).
	Code bool
	// Link marks the run as a hyperlink; Href is the target.
	Link bool
	// Href is the link URL when Link is true.
	Href string
}

// TextColor is a run color: either a soul-bound token role (the documented
// default) or a literal RGB escape hatch. The zero value is the token
// "primary". JSON: { "token": "<role>" } or { "literal": "RRGGBB" }.
type TextColor struct {
	// Token is a semantic text-color role resolved against the soul/theme.
	Token TextColorRole `json:"token,omitempty"`
	// Literal is an explicit "RRGGBB" hex color bypassing the theme.
	Literal string `json:"literal,omitempty"`
}

// TextColorRole names a semantic text-color role (mirrors pptx-go's
// TextColorRole; CONVENTIONS §4 lists the wire values verbatim).
type TextColorRole string

// Text color roles (wire values per CONVENTIONS §4).
const (
	TextPrimary   TextColorRole = "primary"
	TextSecondary TextColorRole = "secondary"
	TextTertiary  TextColorRole = "tertiary"
	TextInverse   TextColorRole = "inverse"
	TextMuted     TextColorRole = "muted"
	TextAccent    TextColorRole = "accent"
	TextAccentAlt TextColorRole = "accentAlt"
	TextSuccess   TextColorRole = "success"
	TextWarning   TextColorRole = "warning"
	TextError     TextColorRole = "error"
)

// TypeRole names a typography scale role (mirrors pptx-go's TypeRole; wire
// values per CONVENTIONS §4: body|h1|h2|h3|code|...).
type TypeRole string

// Typography roles (mirror the define-a-theme skill enum verbatim).
const (
	TypeDisplay   TypeRole = "display"
	TypeH1        TypeRole = "h1"
	TypeH2        TypeRole = "h2"
	TypeH3        TypeRole = "h3"
	TypeH4        TypeRole = "h4"
	TypeH5        TypeRole = "h5"
	TypeBody      TypeRole = "body"
	TypeBodySmall TypeRole = "bodySmall"
	TypeCaption   TypeRole = "caption"
	TypeMono      TypeRole = "mono"
	TypeCode      TypeRole = "code"
)

// MarshalJSON flattens RunStyle and omits an unset Color, producing the
// CONVENTIONS §4 run shape: { text, typeRole?, bold?, ..., color? }.
func (r TextRun) MarshalJSON() ([]byte, error) {
	type plain struct {
		Text      string     `json:"text"`
		TypeRole  TypeRole   `json:"typeRole,omitempty"`
		Bold      bool       `json:"bold,omitempty"`
		Italic    bool       `json:"italic,omitempty"`
		Underline bool       `json:"underline,omitempty"`
		Strike    bool       `json:"strike,omitempty"`
		Code      bool       `json:"code,omitempty"`
		Link      bool       `json:"link,omitempty"`
		Href      string     `json:"href,omitempty"`
		Color     *TextColor `json:"color,omitempty"`
	}
	p := plain{
		Text:      r.Text,
		TypeRole:  r.Style.TypeRole,
		Bold:      r.Style.Bold,
		Italic:    r.Style.Italic,
		Underline: r.Style.Underline,
		Strike:    r.Style.Strike,
		Code:      r.Style.Code,
		Link:      r.Style.Link,
		Href:      r.Style.Href,
	}
	if r.Color != (TextColor{}) {
		c := r.Color
		p.Color = &c
	}
	return json.Marshal(p)
}

// UnmarshalJSON is the inverse of MarshalJSON: it re-folds the flat run
// fields into Style and reads color when present.
func (r *TextRun) UnmarshalJSON(data []byte) error {
	type plain struct {
		Text      string     `json:"text"`
		TypeRole  TypeRole   `json:"typeRole,omitempty"`
		Bold      bool       `json:"bold,omitempty"`
		Italic    bool       `json:"italic,omitempty"`
		Underline bool       `json:"underline,omitempty"`
		Strike    bool       `json:"strike,omitempty"`
		Code      bool       `json:"code,omitempty"`
		Link      bool       `json:"link,omitempty"`
		Href      string     `json:"href,omitempty"`
		Color     *TextColor `json:"color,omitempty"`
	}
	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	r.Text = p.Text
	r.Style = RunStyle{
		TypeRole:  p.TypeRole,
		Bold:      p.Bold,
		Italic:    p.Italic,
		Underline: p.Underline,
		Strike:    p.Strike,
		Code:      p.Code,
		Link:      p.Link,
		Href:      p.Href,
	}
	if p.Color != nil {
		r.Color = *p.Color
	}
	return nil
}
