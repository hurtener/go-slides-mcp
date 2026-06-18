package ir

import (
	"encoding/json"
	"fmt"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// Path addresses a node location in a slide's node tree as an even-length list
// of (field-leg, index) pairs. The first leg is always "nodes" (the slide's
// top-level list); each subsequent leg is a container's child-slice field —
// two_column→"left"/"right", grid→"cells", card→"body", card_section→"body".
// Examples: ["nodes",2] · ["nodes",0,"left",1] · ["nodes",1,"body",0,"cells",3].
// Indices may arrive as int or, from JSON, float64/json.Number. CONVENTIONS §5.
type Path = []any

// Resolve returns the node addressed by p.
func Resolve(s *contracts.Slide, p Path) (contracts.SlideNode, error) {
	slice, idx, err := locate(s, p)
	if err != nil {
		return nil, err
	}
	if idx < 0 || idx >= len(*slice) {
		return nil, fmt.Errorf("ir: index %d out of range (len %d)", idx, len(*slice))
	}
	return (*slice)[idx], nil
}

// Set replaces the node addressed by p.
func Set(s *contracts.Slide, p Path, n contracts.SlideNode) error {
	slice, idx, err := locate(s, p)
	if err != nil {
		return err
	}
	if idx < 0 || idx >= len(*slice) {
		return fmt.Errorf("ir: index %d out of range (len %d)", idx, len(*slice))
	}
	(*slice)[idx] = n
	return nil
}

// Insert inserts n at the index addressed by p (idx == len appends).
func Insert(s *contracts.Slide, p Path, n contracts.SlideNode) error {
	slice, idx, err := locate(s, p)
	if err != nil {
		return err
	}
	if idx < 0 || idx > len(*slice) {
		return fmt.Errorf("ir: insert index %d out of range (len %d)", idx, len(*slice))
	}
	*slice = insertAt(*slice, idx, n)
	return nil
}

// Remove deletes and returns the node addressed by p.
func Remove(s *contracts.Slide, p Path) (contracts.SlideNode, error) {
	slice, idx, err := locate(s, p)
	if err != nil {
		return nil, err
	}
	if idx < 0 || idx >= len(*slice) {
		return nil, fmt.Errorf("ir: index %d out of range (len %d)", idx, len(*slice))
	}
	removed := (*slice)[idx]
	*slice = append((*slice)[:idx], (*slice)[idx+1:]...)
	return removed, nil
}

// Duplicate inserts a deep copy of the node addressed by p immediately after it,
// returning the new node. The copy is via JSON round-trip through the codec.
func Duplicate(s *contracts.Slide, p Path) (contracts.SlideNode, error) {
	slice, idx, err := locate(s, p)
	if err != nil {
		return nil, err
	}
	if idx < 0 || idx >= len(*slice) {
		return nil, fmt.Errorf("ir: index %d out of range (len %d)", idx, len(*slice))
	}
	clone, err := cloneNode((*slice)[idx])
	if err != nil {
		return nil, err
	}
	*slice = insertAt(*slice, idx+1, clone)
	return clone, nil
}

// Move relocates the node at from so it lands at index `to` in the destination
// slice (final-position semantics: after the move, the node is at index toIdx).
// from and to are addressed against the original tree.
func Move(s *contracts.Slide, from, to Path) error {
	fromSlice, fromIdx, err := locate(s, from)
	if err != nil {
		return fmt.Errorf("ir: move from: %w", err)
	}
	if fromIdx < 0 || fromIdx >= len(*fromSlice) {
		return fmt.Errorf("ir: move from index %d out of range (len %d)", fromIdx, len(*fromSlice))
	}
	node := (*fromSlice)[fromIdx]

	toSlice, toIdx, err := locate(s, to)
	if err != nil {
		return fmt.Errorf("ir: move to: %w", err)
	}

	// Remove from the source. When from and to share a parent, toSlice aliases
	// fromSlice, so len(*toSlice) below already reflects the removal.
	*fromSlice = append((*fromSlice)[:fromIdx], (*fromSlice)[fromIdx+1:]...)
	if toIdx < 0 || toIdx > len(*toSlice) {
		*fromSlice = insertAt(*fromSlice, fromIdx, node) // restore on invalid destination
		return fmt.Errorf("ir: move to index %d out of range (len %d)", toIdx, len(*toSlice))
	}
	*toSlice = insertAt(*toSlice, toIdx, node)
	return nil
}

// locate walks p and returns the parent slice + final index (the final index is
// NOT bounds-checked here; each op applies its own rule, e.g. Insert allows len).
func locate(s *contracts.Slide, p Path) (*[]contracts.SlideNode, int, error) {
	if len(p) == 0 || len(p)%2 != 0 {
		return nil, 0, fmt.Errorf("ir: path must be a non-empty even-length list, got %d legs", len(p))
	}
	if leg, _ := p[0].(string); leg != "nodes" {
		return nil, 0, fmt.Errorf("ir: path must start at \"nodes\", got %v", p[0])
	}
	cur := &s.Nodes
	var idx int
	for i := 0; i < len(p); i += 2 {
		var err error
		if idx, err = coerceIndex(p[i+1]); err != nil {
			return nil, 0, err
		}
		if i+2 >= len(p) {
			return cur, idx, nil // final pair
		}
		if idx < 0 || idx >= len(*cur) {
			return nil, 0, fmt.Errorf("ir: index %d out of range at %q (len %d)", idx, p[i], len(*cur))
		}
		nextLeg, ok := p[i+2].(string)
		if !ok {
			return nil, 0, fmt.Errorf("ir: expected a field name at path position %d, got %v", i+2, p[i+2])
		}
		child, err := childSlice((*cur)[idx], nextLeg)
		if err != nil {
			return nil, 0, err
		}
		cur = child
	}
	return cur, idx, nil
}

// childSlice returns a pointer to a container node's named child slice.
func childSlice(n contracts.SlideNode, leg string) (*[]contracts.SlideNode, error) {
	switch v := n.(type) {
	case *contracts.TwoColumn:
		switch leg {
		case "left":
			return &v.Left, nil
		case "right":
			return &v.Right, nil
		}
	case *contracts.Grid:
		if leg == "cells" {
			return &v.Cells, nil
		}
	case *contracts.Card:
		if leg == "body" {
			return &v.Body, nil
		}
	case *contracts.CardSection:
		if leg == "body" {
			return &v.Body, nil
		}
	}
	return nil, fmt.Errorf("ir: node %T has no child slice %q", n, leg)
}

func insertAt(s []contracts.SlideNode, idx int, n contracts.SlideNode) []contracts.SlideNode {
	s = append(s, nil)
	copy(s[idx+1:], s[idx:])
	s[idx] = n
	return s
}

func coerceIndex(v any) (int, error) {
	switch x := v.(type) {
	case int:
		return x, nil
	case int64:
		return int(x), nil
	case float64:
		return int(x), nil
	case json.Number:
		i, err := x.Int64()
		if err != nil {
			return 0, fmt.Errorf("ir: bad index %v: %w", v, err)
		}
		return int(i), nil
	default:
		return 0, fmt.Errorf("ir: expected an integer index, got %T (%v)", v, v)
	}
}

func cloneNode(n contracts.SlideNode) (contracts.SlideNode, error) {
	b, err := json.Marshal(n)
	if err != nil {
		return nil, fmt.Errorf("ir: clone marshal: %w", err)
	}
	return contracts.UnmarshalSlideNode(b)
}
