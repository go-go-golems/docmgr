package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"text/template"
	"time"
)

// CommonTemplateData is the common envelope available to all verb templates
type CommonTemplateData struct {
	Verbs    []string               // Full verb path, e.g., ["docmgr", "doc", "list"]
	Root     string                 // Absolute docs root used
	Now      time.Time              // Rendering timestamp
	Settings map[string]interface{} // Parsed layer values relevant to the verb
}

// RenderVerbTemplate renders a postfix template for a verb if it exists.
// verbPathCandidates is a list of possible verb paths to try (e.g., ["doc", "list"] or ["list", "docs"]).
// Returns true if a template was found and rendered, false otherwise.
// Errors are printed to stderr and are non-fatal.
func RenderVerbTemplate(
	verbPathCandidates [][]string,
	root string,
	settings map[string]interface{},
	data interface{},
) bool {
	// Try each verb path candidate until we find a template
	var templatePath string
	var verbs []string
	for _, candidateVerbs := range verbPathCandidates {
		// Build full verb path including "docmgr" root
		fullVerbs := append([]string{"docmgr"}, candidateVerbs...)
		path := resolveTemplatePath(root, fullVerbs)
		if path != "" {
			templatePath = path
			verbs = fullVerbs
			break
		}
	}

	if templatePath == "" {
		return false
	}

	// Build common envelope
	common := CommonTemplateData{
		Verbs:    verbs,
		Root:     root,
		Now:      time.Now(),
		Settings: settings,
	}

	// Read template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to read template %s: %v\n", templatePath, err)
		return false
	}

	// Create template with FuncMap
	tmpl := template.New(filepath.Base(templatePath)).Funcs(GetTemplateFuncMap())

	// Parse template
	tmpl, err = tmpl.Parse(string(templateContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse template %s: %v\n", templatePath, err)
		return false
	}

	// Combine common envelope with verb-specific data
	templateData := map[string]interface{}{
		"Verbs":    common.Verbs,
		"Root":     common.Root,
		"Now":      common.Now,
		"Settings": common.Settings,
	}

	// Merge verb-specific data (data may be a struct or map)
	if data != nil {
		if dataMap, ok := data.(map[string]interface{}); ok {
			for k, v := range dataMap {
				templateData[k] = v
			}
		} else {
			// If data is a struct, add it as "Data" field
			templateData["Data"] = data
		}
	}

	// Render template to stdout (add newline separator for readability)
	fmt.Fprintln(os.Stdout)
	if err := tmpl.Execute(os.Stdout, templateData); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to render template %s: %v\n", templatePath, err)
		return false
	}

	return true
}

// resolveTemplatePath computes the canonical template path for a verb.
// Returns the path if the template exists, empty string otherwise.
func resolveTemplatePath(root string, verbs []string) string {
	if len(verbs) == 0 {
		return ""
	}

	// Skip root command (e.g., "docmgr")
	// For ["docmgr", "doc", "list"], use templates/doc/list.templ
	// For ["docmgr", "list", "tickets"], use templates/list/tickets.templ
	// For ["docmgr", "doctor"], use templates/doctor.templ

	// Remove "docmgr" root if present
	verbPath := verbs
	if len(verbs) > 0 && verbs[0] == "docmgr" {
		verbPath = verbs[1:]
	}

	if len(verbPath) == 1 {
		// Single-level verb: templates/$verb.templ
		path := filepath.Join(root, "templates", verbPath[0]+".templ")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		return ""
	}

	if len(verbPath) >= 2 {
		// Grouped verb: templates/$group/$verb.templ
		group := verbPath[len(verbPath)-2]
		verb := verbPath[len(verbPath)-1]
		path := filepath.Join(root, "templates", group, verb+".templ")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		return ""
	}

	return ""
}

// GetTemplateFuncMap returns a safe, minimal FuncMap for template rendering
func GetTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"slice": func(start, end int, s interface{}) interface{} {
			switch v := s.(type) {
			case []interface{}:
				if start < 0 {
					start = 0
				}
				if end > len(v) {
					end = len(v)
				}
				if start >= end {
					return []interface{}{}
				}
				return v[start:end]
			case []string:
				if start < 0 {
					start = 0
				}
				if end > len(v) {
					end = len(v)
				}
				if start >= end {
					return []string{}
				}
				return v[start:end]
			default:
				return []interface{}{}
			}
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			result := make(map[string]interface{})
			for i := 0; i < len(values)-1; i += 2 {
				if key, ok := values[i].(string); ok {
					result[key] = values[i+1]
				}
			}
			return result
		},
		"set": func(m map[string]interface{}, key string, value interface{}) map[string]interface{} {
			if m == nil {
				m = make(map[string]interface{})
			}
			m[key] = value
			return m
		},
		"get": func(m map[string]interface{}, key string) interface{} {
			if m == nil {
				return nil
			}
			return m[key]
		},
		"add1": func(n interface{}) int {
			switch v := n.(type) {
			case int:
				return v + 1
			case int64:
				return int(v) + 1
			case float64:
				return int(v) + 1
			default:
				return 1
			}
		},
		"countBy": func(slice interface{}, value interface{}) int {
			count := 0
			valueStr := fmt.Sprintf("%v", value)
			switch v := slice.(type) {
			case []interface{}:
				for _, item := range v {
					if itemMap, ok := item.(map[string]interface{}); ok {
						// Check common fields like "Severity", "Status", etc.
						if severity, ok := itemMap["Severity"]; ok && fmt.Sprintf("%v", severity) == valueStr {
							count++
						} else if status, ok := itemMap["Status"]; ok && fmt.Sprintf("%v", status) == valueStr {
							count++
						}
					} else if fmt.Sprintf("%v", item) == valueStr {
						count++
					}
				}
			case []map[string]interface{}:
				for _, item := range v {
					if severity, ok := item["Severity"]; ok && fmt.Sprintf("%v", severity) == valueStr {
						count++
					} else if status, ok := item["Status"]; ok && fmt.Sprintf("%v", status) == valueStr {
						count++
					}
				}
			default:
				// Try reflection for structs with Severity field
				sliceVal := reflect.ValueOf(slice)
				if sliceVal.Kind() == reflect.Slice {
					for i := 0; i < sliceVal.Len(); i++ {
						item := sliceVal.Index(i)
						if item.Kind() == reflect.Struct {
							severityField := item.FieldByName("Severity")
							if severityField.IsValid() && fmt.Sprintf("%v", severityField.Interface()) == valueStr {
								count++
							}
						}
					}
				}
			}
			return count
		},
	}
}
