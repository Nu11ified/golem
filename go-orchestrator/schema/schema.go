package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type UIComponent struct {
	Type     string                 `json:"type"`
	Props    map[string]interface{} `json:"props,omitempty"`
	Children []UIComponent          `json:"children,omitempty"`
}

type UISchema struct {
	Root UIComponent `json:"root"`
}

func GetExampleSchema() ([]byte, error) {
	schema := UISchema{
		Root: UIComponent{
			Type: "Screen",
			Props: map[string]interface{}{
				"title": "Welcome",
			},
			Children: []UIComponent{
				{
					Type: "Text",
					Props: map[string]interface{}{
						"value": "Hello, user!",
					},
				},
				{
					Type: "Button",
					Props: map[string]interface{}{
						"label":  "Click me",
						"action": "doSomething",
					},
				},
			},
		},
	}
	return json.MarshalIndent(schema, "", "  ")
}

// GetSampleUI returns a sample UI data structure matching the JSON Schema
func GetSampleUI() (map[string]interface{}, error) {
	ui := map[string]interface{}{
		"type": "Screen",
		"props": map[string]interface{}{
			"title": "Welcome",
		},
		"children": []interface{}{
			map[string]interface{}{
				"type": "Text",
				"props": map[string]interface{}{
					"value": "Hello, user!",
				},
			},
			map[string]interface{}{
				"type": "Button",
				"props": map[string]interface{}{
					"label":  "Click me",
					"action": "doSomething",
				},
			},
			map[string]interface{}{
				"type": "Input",
				"props": map[string]interface{}{
					"name":      "email",
					"label":     "Email",
					"inputType": "email",
				},
			},
			map[string]interface{}{
				"type": "List",
				"props": map[string]interface{}{
					"items": []string{"Item 1", "Item 2", "Item 3"},
				},
			},
		},
	}
	return ui, nil
}

// GetUISchemaJSON returns the raw JSON Schema bytes for the UI schema
func GetUISchemaJSON() ([]byte, error) {
	defaultSchemaPath := filepath.Join("..", "node-renderer", "ui-schema.json")
	return os.ReadFile(defaultSchemaPath)
}
