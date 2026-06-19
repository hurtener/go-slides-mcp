# pptx-go engine type reference (READ THIS — do NOT read the module cache)

> Exact Go signatures for github.com/hurtener/pptx-go/scene (the render target) +
> the key pptx theme types. The compose-a-scene + define-a-theme skills are the
> authoritative catalog; this file is the literal struct/enum/func source-of-truth
> so the render adapter never needs the module cache. Render via scene.Render.

## scene package — full type, const, and exported func blocks
```go
type AssetID string
type AssetResolver interface {
	Resolve(ctx context.Context, id AssetID) ([]byte, string, error)
}

func URIAssetResolver(fn func(uuid string) ([]byte, string, error)) AssetResolver {
func TestWithIconExtension_Valid(t *testing.T) {
func TestWithIconExtension_Invalid(t *testing.T) {
func TestValidateIcon(t *testing.T) {
func TestCuratedIconsValidateAtSceneLayer(t *testing.T) {
type LayoutMap map[LayoutKind]string
func DefaultLayoutMap() LayoutMap {
type NodeKind int
const (
	KindHero NodeKind = iota
	KindProse
	KindHeading
	KindList
	KindDivider
	KindQuote
	KindCallout
	KindImage
	KindChip
	KindArrow
	KindCodeBlock
	KindChart
	KindTable
	KindFlow
	KindDecoration
	KindSectionDivider
	KindTwoColumn
	KindGrid
	KindCard
	KindCardSection
)

type SlideNode interface {
	NodeKind() NodeKind
	isSlideNode()
}

type Hero struct {
	node
	Eyebrow  string
	Title    string
	Subtitle string
}

type Prose struct {
	node
	Paragraphs []RichText
}

type Heading struct {
	node
	Text  RichText
	Level int
}

type ListKind int
const (
	ListBullet ListKind = iota
	ListNumber
	ListChecklist
)

type ListItem struct {
	Text    RichText
	Level   int
	Checked bool // checklist items
}

type List struct {
	node
	Kind  ListKind
	Items []ListItem
}

type Divider struct {
	node
	Spacing SpaceRole
}

type Quote struct {
	node
	Text        RichText
	Attribution string
}

type CalloutKind int
const (
	CalloutNote CalloutKind = iota
	CalloutWarning
	CalloutTip
	CalloutImportant
)

type Callout struct {
	node
	Kind  CalloutKind
	Title string
	Body  RichText
}

type FrameKind int
const (
	FrameNone FrameKind = iota
	FrameBrowser
	FramePhone
	FrameDesktop
	FrameLaptop
)

type Crop = pptx.Crop
type Fit = pptx.Fit
const (
	// FitFill stretches the image to fill its box (the zero value / default).
	FitFill = pptx.FitFill
	// FitNone places the image without a stretch fill mode.
	FitNone = pptx.FitNone
)

type Image struct {
	node
	AssetID   AssetID
	Alt       string
	Frame     FrameKind
	FrameName string
	Crop      Crop
	Fit       Fit
}

type ChipTone int
const (
	ChipTint ChipTone = iota
	ChipSolid
	ChipOutline
)

type Chip struct {
	node
	Label string
	Tone  ChipTone
	Color ColorRole
}

type ArrowDirection int
const (
	ArrowRight ArrowDirection = iota
	ArrowLeft
	ArrowUp
	ArrowDown
)

type Arrow struct {
	node
	Direction ArrowDirection
	Label     string
}

type CodeBlock struct {
	node
	AssetID  AssetID
	Language string
	Caption  string
}

type Chart struct {
	node
	AssetID AssetID
	Caption string
}

type Table struct {
	node
	Headers []RichText
	Rows    [][]RichText
	Caption string
}

type FlowOrientation int
const (
	FlowHorizontal FlowOrientation = iota
	FlowVertical
)

type ConnectorKind int
const (
	ConnectorArrow       ConnectorKind = iota // solid arrow (default)
	ConnectorArrowDashed                      // dashed line + chevron head
	ConnectorCycle                            // arrows + a trailing return arrow
	ConnectorPlus                             // a mathPlus glyph between steps
)

type FlowStep struct {
	Label  RichText
	Detail RichText
	Icon   string
}

type Flow struct {
	node
	Orientation FlowOrientation
	Steps       []FlowStep
	Connector   ConnectorKind
}

type DecorationKind int
const (
	// DecorationPreset renders a curated ornament natively (SVG → preset/path).
	DecorationPreset DecorationKind = iota
	// DecorationAsset renders caller-supplied bytes as a pic.
	DecorationAsset
)

type Layer int
const (
	LayerBackground Layer = iota
	LayerForeground
)

type Decoration struct {
	node
	Kind     DecorationKind
	Preset   string // curated ornament name (Kind == DecorationPreset)
	AssetID  AssetID
	Layer    Layer
	Anchor   Anchor
	Offset   Position // EMU shift from the anchor point
	Size     Size     // ornament box; zero = a default size
	Bleed    bool     // allow the box to extend past the slide edge
	Opacity  float64  // 0..1; 0 = fully opaque
	Rotation float64  // degrees clockwise
}

type SectionDivider struct {
	node
	Eyebrow string
	Label   string
}

type ColumnRatio int
const (
	Ratio11 ColumnRatio = iota // 1:1
	Ratio12                    // 1:2
	Ratio21                    // 2:1
)

type TwoColumn struct {
	node
	Ratio ColumnRatio
	Left  []SlideNode
	Right []SlideNode
}

type Grid struct {
	node
	Columns int
	Ratio   []int // per-column weights; empty = equal
	Gap     SpaceRole
	Cells   []SlideNode
}

type BodyLayout int
const (
	BodyVertical BodyLayout = iota
	BodyHorizontal
)

type BorderStyle int
const (
	BorderDefault BorderStyle = iota // defer to Outline
	BorderNone                       // no border (even if Outline is true)
	BorderSolid                      // neutral hairline border
	BorderAccent                     // accent-colored border
)

type CardSize int
const (
	CardSizeMD CardSize = iota
	CardSizeSM
	CardSizeLG
)

type CardLayout int
const (
	CardLayoutDefault CardLayout = iota // icon left of the eyebrow/header stack
	CardLayoutIconTop                   // icon above the eyebrow/header stack
)

type Card struct {
	node
	Header      string
	Eyebrow     string // kicker label above the header
	Icon        string // curated/extension icon name (closed-name; Stage-1 validated)
	HeaderPill  string // pill badge text, right of the header row
	Body        []SlideNode
	BodyLayout  BodyLayout
	Fill        ColorRole
	Outline     bool        // legacy border shorthand; see BorderStyle (D-043)
```
