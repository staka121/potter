package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/staka121/tsubo/pkg/types"
	"gopkg.in/yaml.v3"
)

// ParseTsuboFile parses a .tsubo.yaml file
func ParseTsuboFile(filePath string) (*types.TsuboDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tsubo file: %w", err)
	}

	var tsubo types.TsuboDefinition
	if err := yaml.Unmarshal(data, &tsubo); err != nil {
		return nil, fmt.Errorf("failed to parse tsubo YAML: %w", err)
	}

	return &tsubo, nil
}

// GetContractsDir returns the directory containing contracts
func GetContractsDir(tsuboFilePath string) string {
	return filepath.Dir(tsuboFilePath)
}

// GetProjectRoot returns the project root directory
func GetProjectRoot(contractsDir string) string {
	// Assume contracts are in poc/contracts, so go up two levels
	return filepath.Join(contractsDir, "..", "..")
}
