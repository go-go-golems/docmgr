package templates

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"gopkg.in/yaml.v3"
)

// PrintSchema renders a simple schema describing the structure of v.
// Supported formats: "json" (default), "yaml".
func PrintSchema(w io.Writer, v any, format string) error {
	s := buildSchema(reflect.ValueOf(v))
	var out []byte
	var err error
	switch format {
	case "yaml", "yml":
		out, err = yaml.Marshal(s)
	default:
		out, err = json.MarshalIndent(s, "", "  ")
	}
	if err != nil {
		return err
	}
	_, _ = w.Write(out)
	if len(out) == 0 || out[len(out)-1] != '\n' {
		_, _ = io.WriteString(w, "\n")
	}
	return nil
}

// buildSchema converts a reflected value into a simple, readable schema.
// The returned value is a map[string]any suitable for JSON/YAML marshalling.
func buildSchema(rv reflect.Value) any {
	if !rv.IsValid() {
		return map[string]any{"type": "null"}
	}
	kind := rv.Kind()
	t := rv.Type()
	// Handle typed nils
	if (kind == reflect.Pointer || kind == reflect.Interface || kind == reflect.Slice || kind == reflect.Map) && rv.IsNil() {
		// Return type info if possible
		switch kind {
		case reflect.Slice, reflect.Array:
			return map[string]any{
				"type":  "array",
				"items": buildSchema(reflect.New(t.Elem()).Elem()),
			}
		case reflect.Map:
			return map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}
		case reflect.Interface, reflect.Pointer:
			return map[string]any{"type": "any"}
		case reflect.Invalid:
			return map[string]any{"type": "null"}
		case reflect.Bool:
			return map[string]any{"type": "boolean"}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return map[string]any{"type": "integer"}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return map[string]any{"type": "integer"}
		case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
			return map[string]any{"type": "number"}
		case reflect.String:
			return map[string]any{"type": "string"}
		case reflect.Struct:
			return map[string]any{"type": "object"}
		case reflect.Chan, reflect.Func, reflect.UnsafePointer:
			return map[string]any{"type": "any"}
		}
	}

	switch kind {
	case reflect.Invalid:
		return map[string]any{"type": "null"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]any{"type": "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return map[string]any{"type": "number"}
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Slice, reflect.Array:
		// Try to infer item type from element type or first element
		var itemSchema any
		if rv.Len() > 0 {
			itemSchema = buildSchema(rv.Index(0))
		} else {
			itemSchema = buildSchema(reflect.New(t.Elem()).Elem())
		}
		return map[string]any{
			"type":  "array",
			"items": itemSchema,
		}
	case reflect.Map:
		props := map[string]any{}
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			keyStr := fmt.Sprintf("%v", k.Interface())
			props[keyStr] = buildSchema(v)
		}
		return map[string]any{
			"type":       "object",
			"properties": props,
		}
	case reflect.Struct:
		props := map[string]any{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			// Exported fields only
			if f.PkgPath != "" {
				continue
			}
			name := f.Name
			// If there's a `json:"foo"` tag, prefer it
			if tag, ok := f.Tag.Lookup("json"); ok && tag != "" {
				// Tag may contain options, split on comma
				name = parseJSONTagName(tag, name)
			}
			props[name] = buildSchema(rv.Field(i))
		}
		return map[string]any{
			"type":       "object",
			"properties": props,
		}
	case reflect.Interface, reflect.Pointer:
		return buildSchema(rv.Elem())
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return map[string]any{"type": "any"}
	}
	return map[string]any{"type": "any"}
}

// parseJSONTagName extracts the field name from a `json:"name,omitempty"` tag.
func parseJSONTagName(tag string, fallback string) string {
	if tag == "-" {
		return fallback
	}
	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' {
			if i == 0 {
				return fallback
			}
			return tag[:i]
		}
	}
	return tag
}
