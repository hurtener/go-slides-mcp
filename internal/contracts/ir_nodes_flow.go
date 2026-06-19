package contracts

import (
	"encoding/json"
	"fmt"
)

// FlowOrientation selects a flow's axis (mirrors pptx-go's
// scene.FlowOrientation).
type FlowOrientation string

// Flow orientations (wire values per compose-a-scene).
const (
	FlowHorizontal FlowOrientation = "horizontal"
	FlowVertical   FlowOrientation = "vertical"
)

// ConnectorKind selects the connector style between flow steps (mirrors
// pptx-go's scene.ConnectorKind; ConnectorArrow is the default).
type ConnectorKind string

// Connector styles (wire values per compose-a-scene).
const (
	ConnectorArrow       ConnectorKind = "arrow"
	ConnectorArrowDashed ConnectorKind = "arrow_dashed"
	ConnectorCycle       ConnectorKind = "cycle"
	ConnectorPlus        ConnectorKind = "plus"
)

// FlowStep is one step in a Flow. Mirror of scene.FlowStep. Icon is a closed-
// name curated/extension icon (Stage-1 validated at render time).
type FlowStep struct {
	// Label is the step heading.
	Label RichText `json:"label,omitempty"`
	// Detail is the step body/description.
	Detail RichText `json:"detail,omitempty"`
	// Icon is a closed-name curated/extension icon name.
	Icon string `json:"icon,omitempty"`
}

// Flow is an ordered sequence of connected steps. Renders as native PPTX
// shapes. Mirror of pptx-go's scene.Flow. At least one step is required
// (validation, later unit).
type Flow struct {
	// Orientation is the flow axis.
	Orientation FlowOrientation `json:"orientation,omitempty"`
	// Steps is the ordered flow steps.
	Steps []FlowStep `json:"steps,omitempty"`
	// Connector is the connector style between steps.
	Connector ConnectorKind `json:"connector,omitempty"`
}

func (Flow) slideNodeKind() Kind { return KindFlow }

// MarshalJSON injects the "flow" kind discriminator via marshalNode. Step
// RichText fields marshal through each TextRun's own MarshalJSON.
func (f *Flow) MarshalJSON() ([]byte, error) { return marshalNode(KindFlow, *f) }

// UnmarshalJSON strict-decodes a Flow, then strict-decodes each FlowStep so
// wrong keys (e.g. {title,body} instead of {label,detail}) are a hard error
// naming the offending key(s) and the correct shape. The injected "kind"
// discriminator is explicitly allowed.
func (f *Flow) UnmarshalJSON(data []byte) error {
	type flowWire struct {
		Orientation FlowOrientation   `json:"orientation,omitempty"`
		Steps       []json.RawMessage `json:"steps,omitempty"`
		Connector   ConnectorKind     `json:"connector,omitempty"`
	}
	var wire flowWire
	if err := strictUnmarshal(data, &wire, "kind"); err != nil {
		return err
	}
	f.Orientation = wire.Orientation
	f.Connector = wire.Connector
	if wire.Steps != nil {
		f.Steps = make([]FlowStep, len(wire.Steps))
		for i, raw := range wire.Steps {
			if err := strictUnmarshal(raw, &f.Steps[i]); err != nil {
				if e := asUnknownFieldError(err); e != nil {
					e.Kind = "FlowStep"
				}
				return fmt.Errorf("steps[%d]: %w", i, err)
			}
		}
	}
	return nil
}

func init() { registerNodeKind(KindFlow, func() SlideNode { return &Flow{} }) }
