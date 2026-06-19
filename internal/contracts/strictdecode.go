package contracts

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// UnknownFieldError is returned when a JSON object carries a key that is not
// declared on the target struct. Kind identifies what is being decoded and
// selects the correct-shape hint in Error(); Unknown lists the rejected keys;
// Allowed lists the accepted keys.
type UnknownFieldError struct {
	// Kind identifies the type being decoded (e.g. "FlowStep", "run").
	Kind string
	// Unknown is the sorted list of unrecognised JSON keys.
	Unknown []string
	// Allowed is the sorted list of accepted JSON keys.
	Allowed []string
}

// Error implements the error interface. It names the unknown key(s), lists the
// allowed keys, and — for known kinds — appends a one-line correct-shape hint
// so the model can fix the payload immediately.
func (e *UnknownFieldError) Error() string {
	msg := fmt.Sprintf("unknown field(s) [%s]; allowed: [%s]",
		quoteCSV(e.Unknown), quoteCSV(e.Allowed))
	switch e.Kind {
	case "FlowStep":
		msg += "; correct shape: {\"label\":<RichText>,\"detail\":<RichText>}"
	case "run":
		msg += "; correct shape: {\"text\":\"x\",\"bold\":true,\"italic\":true} (no nested \"style\")"
	}
	return msg
}

func quoteCSV(ss []string) string {
	parts := make([]string, len(ss))
	for i, s := range ss {
		parts[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(parts, ", ")
}

// strictUnmarshal decodes data into v, returning an *UnknownFieldError if data
// contains any JSON object key that is not present in v's struct json-tag set.
// Embedded (anonymous) fields are flattened recursively; tags "-" are skipped;
// the ,omitempty and other options are stripped.
//
// allowExtra names additional keys that are legitimate but not struct fields
// (typically the injected "kind" discriminator). On success the normal
// json.Unmarshal path is taken.
func strictUnmarshal(data []byte, v any, allowExtra ...string) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	allowed := structAllowedKeys(reflect.TypeOf(v))
	for _, k := range allowExtra {
		allowed[k] = true
	}
	var unknown []string
	for k := range raw {
		if !allowed[k] {
			unknown = append(unknown, k)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		keys := make([]string, 0, len(allowed))
		for k := range allowed {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return &UnknownFieldError{Unknown: unknown, Allowed: keys}
	}
	return json.Unmarshal(data, v)
}

// structAllowedKeys returns the set of JSON object keys accepted by t's
// struct type. Embedded (anonymous) fields without an explicit tag are
// recursed into. Fields tagged "-" are excluded. Options (,omitempty etc.)
// are stripped.
func structAllowedKeys(t reflect.Type) map[string]bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return map[string]bool{}
	}
	keys := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if name == "-" {
			continue
		}
		if f.Anonymous && name == "" {
			// Embedded struct without explicit tag: flatten.
			for k := range structAllowedKeys(f.Type) {
				keys[k] = true
			}
			continue
		}
		if name == "" {
			name = f.Name
		}
		keys[name] = true
	}
	return keys
}

// asUnknownFieldError extracts an *UnknownFieldError from err via errors.As.
// Returns nil if err does not wrap one.
func asUnknownFieldError(err error) *UnknownFieldError {
	var e *UnknownFieldError
	if errors.As(err, &e) {
		return e
	}
	return nil
}
