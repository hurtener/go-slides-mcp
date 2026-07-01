package contracts

// LogoToneKind selects a LogoWall's uniform recolor treatment (R14.7,
// D-125). The product mirror of pptx-go's scene.LogoToneKind (an int enum)
// is a string enum per the "string enum, not int" convention.
type LogoToneKind string

// LogoWall tones (wire values). The empty string is accepted (acceptEmpty)
// and maps to LogoToneNone at render time, mirroring the engine's zero
// value — every logo keeps its natural colors.
const (
	// LogoToneNone keeps each logo's natural colors.
	LogoToneNone LogoToneKind = ""
	// LogoToneMono recolors every logo to a brand-neutral two-tone so a
	// mixed set reads as one cohesive, monochrome wall.
	LogoToneMono LogoToneKind = "mono"
	// LogoToneBrand recolors every logo to the accent two-tone.
	LogoToneBrand LogoToneKind = "brand"
)

// LogoWall is an N-up grid of logo assets normalized to a common cell,
// optionally recolored to a uniform tone so a mixed-style set reads as one
// cohesive wall (R14.7, D-125). Each logo is contained (not cropped) and
// centered in its cell. Asset-bearing (resolved via the AssetResolver); a
// missing logo warns and is skipped. Mirror of pptx-go's scene.LogoWall. A
// deck with no LogoWall is byte-identical (a new node, absent until used).
type LogoWall struct {
	// Logos are the wall's logo entries, in display order.
	Logos []LogoEntry `json:"logos,omitempty"`
	// Columns is the number of logos per row (>=1; 0 defaults to a pinned
	// column count at render time).
	Columns int `json:"columns,omitempty"`
	// Tone selects the wall's uniform recolor treatment: "" (none) |
	// "mono" | "brand". Empty = each logo keeps its natural colors.
	Tone LogoToneKind `json:"tone,omitempty"`
	// Caption is an optional heading ("Trusted by", "Integrates with").
	Caption string `json:"caption,omitempty"`
}

func (LogoWall) slideNodeKind() Kind { return KindLogoWall }

// MarshalJSON injects the "logo_wall" kind discriminator via marshalNode.
func (l *LogoWall) MarshalJSON() ([]byte, error) { return marshalNode(KindLogoWall, *l) }

func init() { registerNodeKind(KindLogoWall, func() SlideNode { return &LogoWall{} }) }

// LogoEntry is one logo in a LogoWall (D-125): an asset reference + alt
// text.
type LogoEntry struct {
	// AssetID references the logo image via the AssetResolver.
	AssetID AssetID `json:"assetId,omitempty"`
	// Alt is the logo's accessible alt text.
	Alt string `json:"alt,omitempty"`
}
