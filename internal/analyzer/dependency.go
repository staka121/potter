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
	Port         int
	IsGateway    bool // True if this is an auto-generated gateway service
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
			Port:         objRef.Runtime.Port,
			IsGateway:    false,
		})
	}

	// Auto-generate API Gateway if there are multiple services
	// This implements Tsubo's philosophy: "壺（Tsubo）= single entry point"
	if len(objects) > 1 {
		gateway := createGatewayObject(objects)
		objects = append(objects, gateway)
	}

	return objects, nil
}

// createGatewayObject creates an API Gateway object that depends on all services
func createGatewayObject(services []ObjectWithDeps) ObjectWithDeps {
	// Collect all service names as dependencies
	var allServiceNames []string
	for _, svc := range services {
		allServiceNames = append(allServiceNames, svc.Name)
	}

	return ObjectWithDeps{
		Name:         "gateway-service",
		Contract:     "", // Gateway has no contract file - it's auto-generated
		Dependencies: allServiceNames,
		Port:         8080, // Standard gateway port
		IsGateway:    true,
	}
}
