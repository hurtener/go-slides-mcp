package contracts

import "encoding/json"

// RichText is an ordered list of styled runs — the inline-text content type
// used across the slide IR (CONVENTIONS §4). On the wire it is a JSON ARRAY of
// FLAT run objects; there is no nested "style" object and every key is
// lowercase. A plain run is just {"text":"hello"}; a styled run inlines its
// flags: {"text":"38% lower","bold":true,"italic":true,"code":true,
// "color":{"token":"accent"}}. So a bold-emphasis phrase is two runs:
// [{"text":"Latency "},{"text":"38% lower","bold":true}].
type RichText []TextRun

// PlainText returns the concatenated run text, with no styling — used for
// thumbnails, labels, and previews.
func (rt RichText) PlainText() string {
	if len(rt) == 0 {
		return ""
	}
	var b []byte
	for _, run := range rt {
		b = append(b, run.Text...)
	}
	return string(b)
}

// TextRun is one styled run of text. Its JSON shape is FLAT (CONVENTIONS §4):
// the typography role and inline-formatting flags are inlined as lowercase
// keys on the run object — there is no nested "style" object. Color is omitted
// when unset (meaning the token "primary"). Examples: {"text":"hello"} or
// {"text":"bold bit","bold":true,"italic":true,"color":{"token":"accent"}}.
type TextRun struct {
	// Text is the literal run content.
	Text string `json:"text"`
	// TypeRole selects a typography scale role (body, h2, ...); empty = body.
	TypeRole TypeRole `json:"typeRole,omitempty"`
	// Bold toggles bold weight.
	Bold bool `json:"bold,omitempty"`
	// Italic toggles italic style.
	Italic bool `json:"italic,omitempty"`
	// Underline toggles underline.
	Underline bool `json:"underline,omitempty"`
	// Strike toggles strikethrough.
	Strike bool `json:"strike,omitempty"`
	// Code marks the run as inline code (mono + tint).
	Code bool `json:"code,omitempty"`
	// Link marks the run as a hyperlink; Href is the target.
	Link bool `json:"link,omitempty"`
	// Href is the link URL when Link is true.
	Href string `json:"href,omitempty"`
	// Color is the run color as {"token":"<role>"} or {"literal":"RRGGBB"};
	// omit it (the zero value) for the default token "primary".
	Color TextColor `json:"color,omitempty"`
}

// Style returns the run's typography role and inline-formatting flags grouped
// as a RunStyle — the form the renderer and layout estimator consume. The wire
// shape stays flat; this is an internal convenience over the flat fields.
func (r TextRun) Style() RunStyle {
	return RunStyle{
		TypeRole:  r.TypeRole,
		Bold:      r.Bold,
		Italic:    r.Italic,
		Underline: r.Underline,
		Strike:    r.Strike,
		Code:      r.Code,
		Link:      r.Link,
		Href:      r.Href,
	}
}

// RunStyle mirrors pptx-go's scene.RunStyle: typography role plus inline
// formatting flags and a link target. It is the grouped, in-memory form of a
// TextRun's flat style fields (see TextRun.Style); it is not a wire shape.
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
		TypeRole:  r.TypeRole,
		Bold:      r.Bold,
		Italic:    r.Italic,
		Underline: r.Underline,
		Strike:    r.Strike,
		Code:      r.Code,
		Link:      r.Link,
		Href:      r.Href,
	}
	if r.Color != (TextColor{}) {
		c := r.Color
		p.Color = &c
	}
	return json.Marshal(p)
}

// UnmarshalJSON is the inverse of MarshalJSON: it re-folds the flat run
// fields into Style and reads color when present. Unknown keys (such as a
// nested "style" object) are a hard error naming the offending key(s) and the
// correct flat shape.
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
	if err := strictUnmarshal(data, &p); err != nil {
		if e := asUnknownFieldError(err); e != nil {
			e.Kind = "run"
		}
		return err
	}
	r.Text = p.Text
	r.TypeRole = p.TypeRole
	r.Bold = p.Bold
	r.Italic = p.Italic
	r.Underline = p.Underline
	r.Strike = p.Strike
	r.Code = p.Code
	r.Link = p.Link
	r.Href = p.Href
	if p.Color != nil {
		r.Color = *p.Color
	}
	return nil
}
