package parser

import (
	"fmt"
	"os"

	"github.com/staka121/potter/pkg/types"
	"gopkg.in/yaml.v3"
)

// ParseArchitectureFile parses a .arch.yaml file
func ParseArchitectureFile(filePath string) (*types.ArchitectureDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read architecture file: %w", err)
	}

	var arch types.ArchitectureDefinition
	if err := yaml.Unmarshal(data, &arch); err != nil {
		return nil, fmt.Errorf("failed to parse architecture YAML: %w", err)
	}

	return &arch, nil
}
