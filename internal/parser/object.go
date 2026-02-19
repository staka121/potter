package parser

import (
	"fmt"
	"os"

	"github.com/staka121/potter/pkg/types"
	"gopkg.in/yaml.v3"
)

// ParseObjectFile parses a .object.yaml file
func ParseObjectFile(filePath string) (*types.ObjectDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read object file: %w", err)
	}

	return ParseObjectYAML(data)
}

// ParseObjectYAML parses object definition from raw YAML bytes
func ParseObjectYAML(data []byte) (*types.ObjectDefinition, error) {
	var object types.ObjectDefinition
	if err := yaml.Unmarshal(data, &object); err != nil {
		return nil, fmt.Errorf("failed to parse object YAML: %w", err)
	}

	return &object, nil
}
