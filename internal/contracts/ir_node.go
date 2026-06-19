package contracts

import (
	"encoding/json"
	"fmt"
)

// SlideNode is the sealed-ish union of every slide IR node kind. The marker
// method slideNodeKind is unexported, so only types in this package satisfy
// the interface; the kind registry is the closed dispatch surface.
// CONVENTIONS §3.
type SlideNode interface {
	slideNodeKind() Kind
}

// KindOf returns a node's Kind discriminator — the exported accessor over the
// unexported marker, used by previews/thumbnails outside this package.
func KindOf(n SlideNode) Kind {
	if n == nil {
		return ""
	}
	return n.slideNodeKind()
}

// nodeRegistry maps each Kind discriminator to a constructor returning a
// fresh, zeroed pointer of the concrete node type. Populated by each node
// file's init().
var nodeRegistry = map[Kind]func() SlideNode{}

// registerNodeKind registers a node constructor under its Kind. Called from
// each node file's init().
func registerNodeKind(k Kind, ctor func() SlideNode) {
	nodeRegistry[k] = ctor
}

// UnmarshalSlideNode decodes one JSON object carrying a "kind" discriminator
// into the concrete SlideNode. An unknown or missing kind is a HARD error
// (never silently dropped). Container child slices re-enter this dispatch
// via unmarshalNodes, so nesting is recursive and unbounded. CONVENTIONS §3.
func UnmarshalSlideNode(data []byte) (SlideNode, error) {
	var peek struct {
		Kind Kind `json:"kind"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return nil, fmt.Errorf("slide node: %w", err)
	}
	if peek.Kind == "" {
		return nil, fmt.Errorf("slide node: missing %q discriminator", "kind")
	}
	ctor, ok := nodeRegistry[peek.Kind]
	if !ok {
		return nil, fmt.Errorf("slide node: unknown kind %q", peek.Kind)
	}
	n := ctor()
	if err := strictUnmarshal(data, n, "kind"); err != nil {
		return nil, fmt.Errorf("slide node %q: %w", peek.Kind, err)
	}
	return n, nil
}

// marshalNode is the single shared marshal path every node's MarshalJSON
// calls (CONVENTIONS §3). It serializes the node's own fields and injects
// the "kind" discriminator. fields is the node's value form (the caller
// passes dereferencedNode), whose pointer-receiver MarshalJSON is therefore
// NOT re-invoked — so the map round-trip below sees only the fields and
// never recurses. Output keys are sorted (canonical JSON, CONVENTIONS §6).
func marshalNode(kind Kind, fields any) ([]byte, error) {
	raw, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	out := make(map[string]json.RawMessage)
	if len(raw) > 2 { // not "{}"
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}
	}
	kb, err := json.Marshal(kind)
	if err != nil {
		return nil, err
	}
	out["kind"] = kb
	return json.Marshal(out)
}

// unmarshalNodes decodes a container's child array by dispatching each
// element through UnmarshalSlideNode (the recursive entry point). A nil
// input yields a nil slice, preserving absence vs. emptiness.
func unmarshalNodes(raws []json.RawMessage) ([]SlideNode, error) {
	if raws == nil {
		return nil, nil
	}
	nodes := make([]SlideNode, len(raws))
	for i, raw := range raws {
		n, err := UnmarshalSlideNode(raw)
		if err != nil {
			return nil, fmt.Errorf("[%d]: %w", i, err)
		}
		nodes[i] = n
	}
	return nodes, nil
}
