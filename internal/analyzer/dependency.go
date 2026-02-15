package analyzer

import (
	"fmt"
	"path/filepath"

	"github.com/staka121/potter/internal/parser"
	"github.com/staka121/potter/pkg/types"
)

// ObjectWithDeps represents an object with its dependencies
type ObjectWithDeps struct {
	Name         string
	Contract     string
	Dependencies []string
}

// AnalyzeDependencies analyzes all objects and extracts their service dependencies
func AnalyzeDependencies(tsubo *types.TsuboDefinition, contractsDir string) ([]ObjectWithDeps, error) {
	var objects []ObjectWithDeps

	for _, objRef := range tsubo.Objects {
		// Resolve contract file path
		contractPath := filepath.Join(contractsDir, objRef.Contract)

		// Parse object contract
		objectDef, err := parser.ParseObjectFile(contractPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse contract %s: %w", objRef.Contract, err)
		}

		// Extract service dependencies (not database dependencies)
		var serviceDeps []string
		for _, dep := range objectDef.Dependencies.Services {
			serviceDeps = append(serviceDeps, dep.Name)
		}

		objects = append(objects, ObjectWithDeps{
			Name:         objectDef.Service.Name,
			Contract:     contractPath,
			Dependencies: serviceDeps,
		})
	}

	return objects, nil
}
