package templates

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSlice(t *testing.T) {
	fnMap := GetTemplateFuncMap()
	sliceFn := fnMap["slice"].(func(int, int, interface{}) interface{})

	tests := []struct {
		name     string
		start    int
		end      int
		input    interface{}
		expected interface{}
	}{
		{
			name:     "slice []interface{} normal range",
			start:    1,
			end:      3,
			input:    []interface{}{"a", "b", "c", "d"},
			expected: []interface{}{"b", "c"},
		},
		{
			name:     "slice []interface{} start negative",
			start:    -1,
			end:      2,
			input:    []interface{}{"a", "b", "c"},
			expected: []interface{}{"a", "b"},
		},
		{
			name:     "slice []interface{} end beyond length",
			start:    1,
			end:      10,
			input:    []interface{}{"a", "b", "c"},
			expected: []interface{}{"b", "c"},
		},
		{
			name:     "slice []interface{} start >= end",
			start:    2,
			end:      2,
			input:    []interface{}{"a", "b", "c"},
			expected: []interface{}{},
		},
		{
			name:     "slice []string normal range",
			start:    0,
			end:      2,
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b"},
		},
		{
			name:     "slice []string empty result",
			start:    0,
			end:      0,
			input:    []string{"a", "b", "c"},
			expected: []string{},
		},
		{
			name:     "slice unsupported type",
			start:    0,
			end:      1,
			input:    123,
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sliceFn(tt.start, tt.end, tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("slice(%d, %d, %v) = %v, want %v", tt.start, tt.end, tt.input, result, tt.expected)
			}
		})
	}
}

func TestDict(t *testing.T) {
	fnMap := GetTemplateFuncMap()
	dictFn := fnMap["dict"].(func(...interface{}) map[string]interface{})

	tests := []struct {
		name     string
		args     []interface{}
		expected map[string]interface{}
	}{
		{
			name:     "dict normal pairs",
			args:     []interface{}{"key1", "value1", "key2", 42},
			expected: map[string]interface{}{"key1": "value1", "key2": 42},
		},
		{
			name:     "dict odd number of args",
			args:     []interface{}{"key1", "value1", "key2"},
			expected: map[string]interface{}{"key1": "value1"},
		},
		{
			name:     "dict non-string key",
			args:     []interface{}{123, "value", "key", "value2"},
			expected: map[string]interface{}{"key": "value2"},
		},
		{
			name:     "dict empty",
			args:     []interface{}{},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dictFn(tt.args...)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("dict(%v) = %v, want %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestSet(t *testing.T) {
	fnMap := GetTemplateFuncMap()
	setFn := fnMap["set"].(func(map[string]interface{}, string, interface{}) map[string]interface{})

	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		value    interface{}
		expected map[string]interface{}
	}{
		{
			name:     "set on existing map",
			m:        map[string]interface{}{"a": 1},
			key:      "b",
			value:    "test",
			expected: map[string]interface{}{"a": 1, "b": "test"},
		},
		{
			name:     "set on nil map",
			m:        nil,
			key:      "key",
			value:    "value",
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name:     "set overwrite existing",
			m:        map[string]interface{}{"key": "old"},
			key:      "key",
			value:    "new",
			expected: map[string]interface{}{"key": "new"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := setFn(tt.m, tt.key, tt.value)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("set(%v, %q, %v) = %v, want %v", tt.m, tt.key, tt.value, result, tt.expected)
			}
		})
	}
}

func TestGet(t *testing.T) {
	fnMap := GetTemplateFuncMap()
	getFn := fnMap["get"].(func(map[string]interface{}, string) interface{})

	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected interface{}
	}{
		{
			name:     "get existing key",
			m:        map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "get non-existent key",
			m:        map[string]interface{}{"key": "value"},
			key:      "missing",
			expected: nil,
		},
		{
			name:     "get from nil map",
			m:        nil,
			key:      "key",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFn(tt.m, tt.key)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("get(%v, %q) = %v, want %v", tt.m, tt.key, result, tt.expected)
			}
		})
	}
}

func TestAdd1(t *testing.T) {
	fnMap := GetTemplateFuncMap()
	add1Fn := fnMap["add1"].(func(interface{}) int)

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{
			name:     "add1 int",
			input:    5,
			expected: 6,
		},
		{
			name:     "add1 int64",
			input:    int64(10),
			expected: 11,
		},
		{
			name:     "add1 float64",
			input:    float64(3.7),
			expected: 4,
		},
		{
			name:     "add1 unsupported type",
			input:    "not a number",
			expected: 1,
		},
		{
			name:     "add1 nil",
			input:    nil,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := add1Fn(tt.input)
			if result != tt.expected {
				t.Errorf("add1(%v) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountBy(t *testing.T) {
	fnMap := GetTemplateFuncMap()
	countByFn := fnMap["countBy"].(func(interface{}, interface{}) int)

	// Test struct for reflection-based counting
	type Finding struct {
		Severity string
		Message  string
	}

	tests := []struct {
		name     string
		slice    interface{}
		value    interface{}
		expected int
	}{
		{
			name: "countBy []interface{} with Severity",
			slice: []interface{}{
				map[string]interface{}{"Severity": "ERROR", "Message": "test1"},
				map[string]interface{}{"Severity": "WARNING", "Message": "test2"},
				map[string]interface{}{"Severity": "ERROR", "Message": "test3"},
			},
			value:    "ERROR",
			expected: 2,
		},
		{
			name: "countBy []interface{} with Status",
			slice: []interface{}{
				map[string]interface{}{"Status": "active", "Title": "test1"},
				map[string]interface{}{"Status": "complete", "Title": "test2"},
				map[string]interface{}{"Status": "active", "Title": "test3"},
			},
			value:    "active",
			expected: 2,
		},
		{
			name: "countBy []map[string]interface{} with Severity",
			slice: []map[string]interface{}{
				{"Severity": "WARNING", "Message": "test1"},
				{"Severity": "WARNING", "Message": "test2"},
				{"Severity": "ERROR", "Message": "test3"},
			},
			value:    "WARNING",
			expected: 2,
		},
		{
			name: "countBy []struct with Severity field (reflection)",
			slice: []Finding{
				{Severity: "ERROR", Message: "test1"},
				{Severity: "WARNING", Message: "test2"},
				{Severity: "ERROR", Message: "test3"},
			},
			value:    "ERROR",
			expected: 2,
		},
		{
			name: "countBy no matches",
			slice: []interface{}{
				map[string]interface{}{"Severity": "WARNING", "Message": "test1"},
				map[string]interface{}{"Severity": "WARNING", "Message": "test2"},
			},
			value:    "ERROR",
			expected: 0,
		},
		{
			name:     "countBy empty slice",
			slice:    []interface{}{},
			value:    "ERROR",
			expected: 0,
		},
		{
			name:     "countBy unsupported type",
			slice:    "not a slice",
			value:    "ERROR",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countByFn(tt.slice, tt.value)
			if result != tt.expected {
				t.Errorf("countBy(%v, %v) = %d, want %d", tt.slice, tt.value, result, tt.expected)
			}
		})
	}
}

func TestResolveTemplatePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		root        string
		verbs       []string
		createFile  bool
		expected    string
		shouldExist bool
	}{
		{
			name:        "single verb",
			root:        tmpDir,
			verbs:       []string{"docmgr", "doctor"},
			createFile:  true,
			expected:    tmpDir + "/templates/doctor.templ",
			shouldExist: true,
		},
		{
			name:        "grouped verb",
			root:        tmpDir,
			verbs:       []string{"docmgr", "doc", "list"},
			createFile:  true,
			expected:    tmpDir + "/templates/doc/list.templ",
			shouldExist: true,
		},
		{
			name:        "grouped verb list tickets",
			root:        tmpDir,
			verbs:       []string{"docmgr", "list", "tickets"},
			createFile:  true,
			expected:    tmpDir + "/templates/list/tickets.templ",
			shouldExist: true,
		},
		{
			name:        "file doesn't exist",
			root:        tmpDir,
			verbs:       []string{"docmgr", "missing"},
			createFile:  false,
			expected:    "",
			shouldExist: false,
		},
		{
			name:        "empty verbs",
			root:        tmpDir,
			verbs:       []string{},
			createFile:  false,
			expected:    "",
			shouldExist: false,
		},
		{
			name:        "only docmgr root",
			root:        tmpDir,
			verbs:       []string{"docmgr"},
			createFile:  false,
			expected:    "",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createFile {
				// Create the directory structure and file
				expectedPath := tt.expected
				dir := filepath.Dir(expectedPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(expectedPath, []byte("test template"), 0644); err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
			}

			result := resolveTemplatePath(tt.root, tt.verbs)
			if tt.shouldExist {
				if result != tt.expected {
					t.Errorf("resolveTemplatePath(%q, %v) = %q, want %q", tt.root, tt.verbs, result, tt.expected)
				}
			} else {
				if result != "" {
					t.Errorf("resolveTemplatePath(%q, %v) = %q, want empty string", tt.root, tt.verbs, result)
				}
			}
		})
	}
}

