package contracts

import (
	"encoding/json"
	"testing"
)

// TestDescribeNodeExamplesRoundTrip asserts that every ExampleNodeForKind
// result round-trips through UnmarshalSlideNode without dropping any field.
// This is the binding acceptance criterion for Phase-12-FIDELITY B1.
func TestDescribeNodeExamplesRoundTrip(t *testing.T) {
	for _, k := range RegisteredKinds() {
		k := k
		t.Run(k, func(t *testing.T) {
			t.Parallel()
			kind := Kind(k)

			node, ok := ExampleNodeForKind(kind)
			if !ok {
				t.Fatalf("ExampleNodeForKind(%q) returned false — add an example", kind)
			}

			// Marshal the original populated node.
			b1, err := json.Marshal(node)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			// Round-trip through the registry dispatcher.
			decoded, err := UnmarshalSlideNode(b1)
			if err != nil {
				t.Fatalf("UnmarshalSlideNode: %v", err)
			}

			// Re-marshal and compare field-by-field via JSON maps so key
			// ordering differences do not cause false failures.
			b2, err := json.Marshal(decoded)
			if err != nil {
				t.Fatalf("re-marshal: %v", err)
			}

			var m1, m2 map[string]any
			if err := json.Unmarshal(b1, &m1); err != nil {
				t.Fatalf("parse original as map: %v", err)
			}
			if err := json.Unmarshal(b2, &m2); err != nil {
				t.Fatalf("parse round-trip as map: %v", err)
			}

			// jsonEqual does a deep string comparison via re-marshal.
			j1, _ := json.Marshal(m1)
			j2, _ := json.Marshal(m2)
			if string(j1) != string(j2) {
				t.Fatalf("round-trip field mismatch for %q:\n  original:  %s\n  round-trip:%s", kind, b1, b2)
			}
		})
	}
}

// TestDescribeNodeFlowExample asserts the flow example uses label+detail steps
// and not the El Mate wrong shape (title/body).
func TestDescribeNodeFlowExample(t *testing.T) {
	node, ok := ExampleNodeForKind(KindFlow)
	if !ok {
		t.Fatal("ExampleNodeForKind(flow) returned false")
	}
	flow, ok := node.(*Flow)
	if !ok {
		t.Fatalf("expected *Flow, got %T", node)
	}
	if len(flow.Steps) == 0 {
		t.Fatal("flow example has no steps")
	}
	for i, s := range flow.Steps {
		if len(s.Label) == 0 {
			t.Errorf("steps[%d].Label is empty (must be non-empty RichText)", i)
		}
		if len(s.Detail) == 0 {
			t.Errorf("steps[%d].Detail is empty (must be non-empty RichText)", i)
		}
	}

	// Verify the JSON shape carries label+detail, not title+body.
	b, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	steps, _ := m["steps"].([]any)
	if len(steps) == 0 {
		t.Fatal("marshaled flow has no steps")
	}
	first, _ := steps[0].(map[string]any)
	if _, ok := first["label"]; !ok {
		t.Errorf("flow step JSON missing 'label' key: %v", first)
	}
	if _, ok := first["detail"]; !ok {
		t.Errorf("flow step JSON missing 'detail' key: %v", first)
	}
	if _, bad := first["title"]; bad {
		t.Errorf("flow step JSON has wrong key 'title' (should be 'label')")
	}
	if _, bad := first["body"]; bad {
		t.Errorf("flow step JSON has wrong key 'body' (should be 'detail')")
	}
}

// TestDescribeNodeCalloutExample asserts the callout example carries calloutKind.
func TestDescribeNodeCalloutExample(t *testing.T) {
	node, ok := ExampleNodeForKind(KindCallout)
	if !ok {
		t.Fatal("ExampleNodeForKind(callout) returned false")
	}
	callout, ok := node.(*Callout)
	if !ok {
		t.Fatalf("expected *Callout, got %T", node)
	}
	if callout.Kind == "" {
		t.Fatal("callout example missing calloutKind")
	}

	b, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["calloutKind"]; !ok {
		t.Errorf("callout JSON missing 'calloutKind': %s", b)
	}
}

// TestDescribeNodeAllKindsCovered asserts that RegisteredKinds and
// ExampleNodeForKind are in sync — every registered kind has an example.
func TestDescribeNodeAllKindsCovered(t *testing.T) {
	for _, k := range RegisteredKinds() {
		if _, ok := ExampleNodeForKind(Kind(k)); !ok {
			t.Errorf("ExampleNodeForKind(%q) returned false — add an example to describe_node_examples.go", k)
		}
	}
}
