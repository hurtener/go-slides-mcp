package contracts

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// TestStrictUnmarshalAllowedKeys verifies that a JSON object whose keys all
// match the target struct's json-tags is decoded without error.
func TestStrictUnmarshalAllowedKeys(t *testing.T) {
	type target struct {
		Name string `json:"name"`
		Age  int    `json:"age,omitempty"`
	}
	var v target
	if err := strictUnmarshal([]byte(`{"name":"alice","age":30}`), &v); err != nil {
		t.Fatalf("strictUnmarshal: unexpected error: %v", err)
	}
	if v.Name != "alice" || v.Age != 30 {
		t.Fatalf("decoded value = %+v, want name=alice age=30", v)
	}
}

// TestStrictUnmarshalUnknownField verifies that an unknown JSON key yields an
// *UnknownFieldError naming the bad key(s) and the allowed set.
func TestStrictUnmarshalUnknownField(t *testing.T) {
	type target struct {
		Name string `json:"name"`
	}
	var v target
	err := strictUnmarshal([]byte(`{"name":"alice","extra":"x"}`), &v)
	if err == nil {
		t.Fatal("strictUnmarshal: want error for unknown field, got nil")
	}
	var ufe *UnknownFieldError
	if !errors.As(err, &ufe) {
		t.Fatalf("error type = %T, want *UnknownFieldError", err)
	}
	if len(ufe.Unknown) != 1 || ufe.Unknown[0] != "extra" {
		t.Errorf("Unknown = %v, want [extra]", ufe.Unknown)
	}
	if len(ufe.Allowed) != 1 || ufe.Allowed[0] != "name" {
		t.Errorf("Allowed = %v, want [name]", ufe.Allowed)
	}
}

// TestStrictUnmarshalAllowExtra verifies that keys listed in allowExtra are
// permitted even though they are not struct fields (e.g. the injected "kind").
func TestStrictUnmarshalAllowExtra(t *testing.T) {
	type target struct {
		Name string `json:"name"`
	}
	var v target
	if err := strictUnmarshal([]byte(`{"name":"alice","kind":"hero"}`), &v, "kind"); err != nil {
		t.Fatalf("strictUnmarshal: unexpected error with allowExtra: %v", err)
	}
}

// TestUnknownFieldErrorFlowStepHint verifies that an *UnknownFieldError for a
// FlowStep includes the {label,detail} correct-shape hint.
func TestUnknownFieldErrorFlowStepHint(t *testing.T) {
	e := &UnknownFieldError{
		Kind:    "FlowStep",
		Unknown: []string{"body", "title"},
		Allowed: []string{"detail", "icon", "label"},
	}
	msg := e.Error()
	for _, want := range []string{"title", "label", "detail"} {
		if !strings.Contains(msg, want) {
			t.Errorf("Error() = %q; want it to contain %q", msg, want)
		}
	}
}

// TestUnknownFieldErrorRunHint verifies that an *UnknownFieldError for a run
// includes the flat-style hint and explicitly calls out "style".
func TestUnknownFieldErrorRunHint(t *testing.T) {
	e := &UnknownFieldError{
		Kind:    "run",
		Unknown: []string{"style"},
		Allowed: []string{"bold", "italic", "text"},
	}
	msg := e.Error()
	for _, want := range []string{"style", "bold"} {
		if !strings.Contains(msg, want) {
			t.Errorf("Error() = %q; want it to contain %q", msg, want)
		}
	}
}

// TestFlowStrictDecodeRejectsWrongStepKeys is the primary acceptance check:
// decoding a flow with {title,body} steps must error, naming those keys.
func TestFlowStrictDecodeRejectsWrongStepKeys(t *testing.T) {
	data := []byte(`{"kind":"flow","steps":[{"title":"a","body":"b"}]}`)
	_, err := UnmarshalSlideNode(data)
	if err == nil {
		t.Fatal("UnmarshalSlideNode: want error for wrong step keys, got nil")
	}
	for _, want := range []string{"title", "label"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q; want it to contain %q", err.Error(), want)
		}
	}
}

// TestTextRunStrictDecodeRejectsNestedStyle is the primary acceptance check:
// decoding a run with a nested "style" object must error naming that key.
func TestTextRunStrictDecodeRejectsNestedStyle(t *testing.T) {
	data := []byte(`{"text":"x","style":{"italic":true}}`)
	var run TextRun
	err := run.UnmarshalJSON(data)
	if err == nil {
		t.Fatal("TextRun.UnmarshalJSON: want error for nested style, got nil")
	}
	var ufe *UnknownFieldError
	if !errors.As(err, &ufe) {
		t.Fatalf("error type = %T, want *UnknownFieldError", err)
	}
	if ufe.Kind != "run" {
		t.Errorf("Kind = %q, want %q", ufe.Kind, "run")
	}
	if !strings.Contains(err.Error(), "style") {
		t.Errorf("error = %q; want it to contain %q", err.Error(), "style")
	}
}

// TestLegitKindKeyDecodes verifies that the injected "kind" discriminator does
// NOT trigger an unknown-field error during default-path decode.
func TestLegitKindKeyDecodes(t *testing.T) {
	data := []byte(`{"kind":"heading","text":[{"text":"Title"}],"level":2}`)
	n, err := UnmarshalSlideNode(data)
	if err != nil {
		t.Fatalf("UnmarshalSlideNode: unexpected error: %v", err)
	}
	h, ok := n.(*Heading)
	if !ok {
		t.Fatalf("node type = %T, want *Heading", n)
	}
	if got := h.Text.PlainText(); got != "Title" {
		t.Errorf("heading text = %q, want %q", got, "Title")
	}
}

// TestListStrictDecodeRejectsUnknownItemKey verifies list item strictness.
func TestListStrictDecodeRejectsUnknownItemKey(t *testing.T) {
	data := []byte(`{"kind":"list","listKind":"bullet","items":[{"content":"x"}]}`)
	_, err := UnmarshalSlideNode(data)
	if err == nil {
		t.Fatal("UnmarshalSlideNode: want error for unknown list item key, got nil")
	}
	if !strings.Contains(err.Error(), "content") {
		t.Errorf("error = %q; want it to contain %q", err.Error(), "content")
	}
}

// TestFlowRoundTripStaysGreen ensures strict-decode does NOT reject a
// correctly-shaped flow node (regression: strictness must not break valid IR).
func TestFlowRoundTripStaysGreen(t *testing.T) {
	node := &Flow{
		Orientation: FlowHorizontal,
		Connector:   ConnectorArrow,
		Steps: []FlowStep{
			{Label: RichText{{Text: "start"}}, Detail: RichText{{Text: "go"}}, Icon: "play"},
			{Label: RichText{{Text: "end"}, {Text: "!", Bold: true}}},
		},
	}
	b, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := UnmarshalSlideNode(b)
	if err != nil {
		t.Fatalf("UnmarshalSlideNode: %v", err)
	}
	f, ok := got.(*Flow)
	if !ok {
		t.Fatalf("node type = %T, want *Flow", got)
	}
	if len(f.Steps) != 2 {
		t.Fatalf("steps count = %d, want 2", len(f.Steps))
	}
	if got := f.Steps[0].Label.PlainText(); got != "start" {
		t.Errorf("step[0].label = %q, want %q", got, "start")
	}
}

// TestListRoundTripStaysGreen ensures strict-decode does NOT reject a
// correctly-shaped list node.
func TestListRoundTripStaysGreen(t *testing.T) {
	node := &List{
		Kind: ListChecklist,
		Items: []ListItem{
			{Text: RichText{{Text: "ship it"}}, Checked: true},
			{Text: RichText{{Text: "nested"}}, Level: 1},
		},
	}
	b, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := UnmarshalSlideNode(b)
	if err != nil {
		t.Fatalf("UnmarshalSlideNode: %v", err)
	}
	l, ok := got.(*List)
	if !ok {
		t.Fatalf("node type = %T, want *List", got)
	}
	if len(l.Items) != 2 {
		t.Fatalf("items count = %d, want 2", len(l.Items))
	}
	if got := l.Items[0].Text.PlainText(); got != "ship it" {
		t.Errorf("items[0].text = %q, want %q", got, "ship it")
	}
}
