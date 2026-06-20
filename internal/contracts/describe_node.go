package contracts

import (
	"encoding/json"
	"sort"
)

// DescribeNodeInput is the typed input for the describe_node tool.
type DescribeNodeInput struct {
	// Kind is the node kind discriminator to describe. Empty string returns
	// every registered kind (the full vocabulary).
	Kind string `json:"kind,omitempty"`
}

// DescribeNodeOutput is the structured result for describe_node.
type DescribeNodeOutput struct {
	// Nodes is the per-kind shape description, one entry per returned kind.
	Nodes []NodeShape `json:"nodes"`
}

// NodeShape is the authoritative shape description for one slide node kind.
type NodeShape struct {
	// Kind is the node kind discriminator — the value of the "kind" JSON field
	// that selects this node type.
	Kind string `json:"kind"`
	// Summary is a one-line agent-facing description of the node's purpose
	// and its key fields.
	Summary string `json:"summary"`
	// Fields is the ordered list of JSON fields for this node kind, derived
	// from the contract struct. The "kind" discriminator itself is omitted
	// (it is always injected by MarshalJSON).
	Fields []NodeField `json:"fields"`
	// Example is a canonical, schema-valid JSON object for this kind, built
	// by marshalling a populated real node struct. Safe to copy/paste directly
	// into a slide IR — it always round-trips through UnmarshalSlideNode.
	Example json.RawMessage `json:"example"`
}

// NodeField describes one JSON field of a node kind.
type NodeField struct {
	// Name is the JSON field name (the json tag key).
	Name string `json:"name"`
	// JSONType is a human-readable description of the field's JSON wire type.
	JSONType string `json:"jsonType"`
	// Required is true when the field has no omitempty tag and must always
	// appear in the JSON object.
	Required bool `json:"required,omitempty"`
	// IsRichText is true when the field's wire type is a RichText value — a
	// JSON array of flat run objects [{text,bold?,italic?,…}]. Do NOT use a
	// plain string for a RichText field, and do NOT nest a style object.
	IsRichText bool `json:"isRichText,omitempty"`
	// Note is an optional clarifying note for agents, surfacing common
	// gotchas or enum values.
	Note string `json:"note,omitempty"`
}

// RegisteredKinds returns every registered node kind discriminator in sorted
// order. The list is sourced from the internal node registry so it always
// matches the set that UnmarshalSlideNode can decode — it cannot drift.
func RegisteredKinds() []string {
	kinds := make([]string, 0, len(nodeRegistry))
	for k := range nodeRegistry {
		kinds = append(kinds, string(k))
	}
	sort.Strings(kinds)
	return kinds
}
