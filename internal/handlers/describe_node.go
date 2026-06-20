package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

var (
	// richTextType is the reflect.Type for contracts.RichText (= []TextRun),
	// used to identify RichText fields during struct reflection.
	richTextType = reflect.TypeOf(contracts.RichText(nil))
	// slideNodeType is the reflect.Type for the contracts.SlideNode interface,
	// used to identify child-node slice fields.
	slideNodeType = reflect.TypeOf((*contracts.SlideNode)(nil)).Elem()
)

// describeNode returns the authoritative JSON shape of the requested node kind
// (or all kinds when Kind is empty). Fields are derived from the contract
// struct via reflection; the Example is a real, populated struct that always
// round-trips through UnmarshalSlideNode.
func (h *handlers) describeNode(_ context.Context, in contracts.DescribeNodeInput) (tool.Result[contracts.DescribeNodeOutput], error) {
	all := contracts.RegisteredKinds()

	var kinds []string
	if in.Kind != "" {
		found := false
		for _, k := range all {
			if k == in.Kind {
				found = true
				break
			}
		}
		if !found {
			msg := fmt.Sprintf("unknown node kind %q; valid kinds: %s", in.Kind, strings.Join(all, ", "))
			return tool.Result[contracts.DescribeNodeOutput]{
				Text:       msg,
				Structured: contracts.DescribeNodeOutput{Nodes: []contracts.NodeShape{}},
			}, nil
		}
		kinds = []string{in.Kind}
	} else {
		kinds = all
	}

	shapes := make([]contracts.NodeShape, 0, len(kinds))
	for _, k := range kinds {
		shape, err := buildNodeShape(contracts.Kind(k))
		if err != nil {
			return tool.Result[contracts.DescribeNodeOutput]{}, fmt.Errorf("describe_node build %q: %w", k, err)
		}
		shapes = append(shapes, shape)
	}

	out := contracts.DescribeNodeOutput{Nodes: shapes}
	return tool.Result[contracts.DescribeNodeOutput]{
		Text:       agentText(fmt.Sprintf("describe_node: %d kind(s)", len(shapes)), out),
		Structured: out,
	}, nil
}

// buildNodeShape builds the NodeShape for one kind using reflection for fields
// and ExampleNodeForKind for the canonical example object.
func buildNodeShape(kind contracts.Kind) (contracts.NodeShape, error) {
	node, ok := contracts.ExampleNodeForKind(kind)
	if !ok {
		return contracts.NodeShape{}, fmt.Errorf("no example registered for kind %q", kind)
	}
	exJSON, err := json.Marshal(node)
	if err != nil {
		return contracts.NodeShape{}, fmt.Errorf("marshal example for %q: %w", kind, err)
	}
	return contracts.NodeShape{
		Kind:    string(kind),
		Summary: nodeShapeSummary(kind),
		Fields:  buildFields(node),
		Example: json.RawMessage(exJSON),
	}, nil
}

// buildFields derives []NodeField from the concrete node type via reflection,
// walking all json-tagged struct fields.
func buildFields(node contracts.SlideNode) []contracts.NodeField {
	t := reflect.TypeOf(node)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var fields []contracts.NodeField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		parts := strings.SplitN(tag, ",", 2)
		name := parts[0]
		if name == "" || name == "-" {
			continue
		}
		omitempty := len(parts) > 1 && strings.Contains(parts[1], "omitempty")
		jType, isRT := jsonTypeLabel(f.Type)

		nf := contracts.NodeField{
			Name:       name,
			JSONType:   jType,
			Required:   !omitempty,
			IsRichText: isRT,
			Note:       fieldNote(name, f.Type),
		}
		fields = append(fields, nf)
	}
	return fields
}

// fieldNote returns an agent-facing clarifying note for fields that are
// commonly misused.
func fieldNote(name string, t reflect.Type) string {
	switch name {
	case "listKind":
		return "list variant: bullet|number|checklist. NOT the node discriminator."
	case "calloutKind":
		return "callout variant: note|warning|tip|important. NOT the node discriminator."
	case "decorationKind":
		return "decoration variant: preset|asset. NOT the node discriminator."
	case "steps":
		return "each step: {label: RichText, detail: RichText, icon?: string}. Do NOT use title/body."
	case "items":
		return "each item: {text: RichText, level?: number (0=top), checked?: boolean}."
	case "body":
		if t.Kind() == reflect.Slice && t.Elem().Implements(slideNodeType) {
			return "array of child slide nodes, each with a kind discriminator."
		}
	case "cells":
		return "array of child slide nodes; length must be a multiple of columns."
	case "headers":
		return "one RichText per column (array of arrays of run objects)."
	case "rows":
		return "one RichText array per cell; each row has one entry per column."
	case "paragraphs":
		return "one RichText per paragraph (array of arrays of run objects)."
	}
	return ""
}

// jsonTypeLabel returns a human-readable JSON type label and an isRichText
// flag for a reflect.Type.
func jsonTypeLabel(t reflect.Type) (string, bool) {
	// Direct RichText named type ([]TextRun).
	if t == richTextType {
		return "RichText ([{text,bold?,italic?,…}])", true
	}
	switch t.Kind() {
	case reflect.String:
		return "string", false
	case reflect.Bool:
		return "boolean", false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number", false
	case reflect.Float32, reflect.Float64:
		return "number", false
	case reflect.Slice:
		elem := t.Elem()
		// []RichText — e.g. Table.Headers (one per column).
		if elem == richTextType {
			return "array of RichText", false
		}
		// [][]RichText — e.g. Table.Rows (one RichText per cell).
		if elem.Kind() == reflect.Slice && elem.Elem() == richTextType {
			return "array of RichText rows ([][]RichText)", false
		}
		// []SlideNode — child node containers.
		if elem.Implements(slideNodeType) {
			return "array of slide nodes", false
		}
		// []int — e.g. Grid.Ratio.
		if elem.Kind() == reflect.Int {
			return "array of number", false
		}
		// []SomeStruct — e.g. []FlowStep, []ListItem.
		if elem.Kind() == reflect.Struct {
			return "array of " + elem.Name(), false
		}
		if elem.Kind() == reflect.Pointer && elem.Elem().Kind() == reflect.Struct {
			return "array of " + elem.Elem().Name(), false
		}
		return "array", false
	case reflect.Struct:
		return "object (" + t.Name() + ")", false
	default:
		return t.String(), false
	}
}

// nodeShapeSummary returns a one-line agent-facing description for each kind.
func nodeShapeSummary(kind contracts.Kind) string {
	m := map[contracts.Kind]string{
		contracts.KindHero:           "Cover/title: eyebrow (string) + title (string) + subtitle (string). Use on cover slides.",
		contracts.KindHeading:        "Heading line: text (RichText) + level 1–6. State the slide takeaway.",
		contracts.KindProse:          "Body paragraphs: paragraphs ([]RichText, one entry per paragraph). Keep to 1–3.",
		contracts.KindList:           "Bullet/number/checklist: listKind + items[].text (RichText). Max 6 items.",
		contracts.KindCallout:        "Highlighted note: calloutKind (note|warning|tip|important) + title (string) + body (RichText).",
		contracts.KindTwoColumn:      "Two-column split: ratio (1:1|1:2|2:1) + left/right each as []SlideNode.",
		contracts.KindGrid:           "Column grid: columns (2–4) + cells ([]SlideNode). Prefer Card children.",
		contracts.KindCard:           "Accent card: header (string) + body ([]SlideNode). Add fill/elevation/eyebrow for design.",
		contracts.KindCardSection:    "Top-level card section: header (string) + body ([]SlideNode, must be non-empty).",
		contracts.KindDivider:        "Horizontal rule with surrounding spacing token role.",
		contracts.KindQuote:          "Pull-quote: text (RichText) + attribution (string).",
		contracts.KindChip:           "Small tag/badge: label (string) + tone (tint|solid|outline) + color role.",
		contracts.KindArrow:          "Directional connector: direction (right|left|up|down) + optional label (string).",
		contracts.KindSectionDivider: "Full-bleed section break: eyebrow (string) + label (string).",
		contracts.KindTable:          "Data table: headers ([]RichText, one per col) + rows ([][]RichText) + caption (string).",
		contracts.KindFlow:           "Process sequence: steps[].label (RichText) + steps[].detail (RichText). NOT title/body.",
		contracts.KindImage:          "Picture: assetId from upload_asset + alt (string) + optional frame/fit.",
		contracts.KindCodeBlock:      "Code listing raster: assetId from compile_code + language (string) + caption (string).",
		contracts.KindChart:          "Chart raster: assetId from compile_chart + caption (string).",
		contracts.KindDecoration:     "Ornament: decorationKind (preset|asset) + preset name or assetId + layer + anchor.",
	}
	if s, ok := m[kind]; ok {
		return s
	}
	return string(kind) + " node"
}
