package templates

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPrintSchema_JSON_MapWithSliceStruct(t *testing.T) {
	type Item struct {
		Name  string
		Count int
	}
	data := map[string]any{
		"Total": 1,
		"Items": []Item{
			{Name: "a", Count: 2},
		},
	}
	var buf bytes.Buffer
	err := PrintSchema(&buf, data, "json")
	if err != nil {
		t.Fatalf("PrintSchema failed: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}
	if m["type"] != "object" {
		t.Fatalf("expected type object, got %v", m["type"])
	}
	props, ok := m["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties object, got %T", m["properties"])
	}
	total, ok := props["Total"].(map[string]any)
	if !ok || total["type"] != "integer" {
		t.Fatalf("expected Total integer schema, got %#v", total)
	}
	items, ok := props["Items"].(map[string]any)
	if !ok || items["type"] != "array" {
		t.Fatalf("expected Items array schema, got %#v", items)
	}
	itemSchema, ok := items["items"].(map[string]any)
	if !ok || itemSchema["type"] != "object" {
		t.Fatalf("expected Items.items object schema, got %#v", itemSchema)
	}
	itemProps, ok := itemSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected Items.items.properties object, got %#v", itemSchema["properties"])
	}
	if name, ok := itemProps["Name"].(map[string]any); !ok || name["type"] != "string" {
		t.Fatalf("expected Name string, got %#v", itemProps["Name"])
	}
	if count, ok := itemProps["Count"].(map[string]any); !ok || count["type"] != "integer" {
		t.Fatalf("expected Count integer, got %#v", itemProps["Count"])
	}
}

func TestPrintSchema_YAML_Simple(t *testing.T) {
	data := map[string]any{
		"A": "x",
		"B": 2,
	}
	var buf bytes.Buffer
	if err := PrintSchema(&buf, data, "yaml"); err != nil {
		t.Fatalf("PrintSchema failed: %v", err)
	}
	s := buf.String()
	if len(s) == 0 || !bytes.Contains([]byte(s), []byte("type: object")) {
		t.Fatalf("expected yaml with type: object, got: %s", s)
	}
}


