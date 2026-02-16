package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/staka121/potter/pkg/types"
)

// Generator handles Kubernetes manifest generation
type Generator struct {
	config *GeneratorConfig
}

// NewGenerator creates a new K8s manifest generator
func NewGenerator(config *GeneratorConfig) *Generator {
	if config == nil {
		config = DefaultGeneratorConfig()
	}
	return &Generator{
		config: config,
	}
}

// Generate generates all Kubernetes manifests from a Tsubo definition
func (g *Generator) Generate(tsuboDef *types.TsuboDefinition) (*ManifestSet, error) {
	manifests := &ManifestSet{
		Namespace:   g.config.Namespace,
		Deployments: make([]string, 0),
		Services:    make([]string, 0),
		ConfigMaps:  make([]string, 0),
	}

	tsuboName := tsuboDef.Tsubo.Name

	// Generate namespace
	namespaceManifest := GenerateNamespace(g.config.Namespace, tsuboName)

	// Generate manifests for each object (service)
	for _, obj := range tsuboDef.Objects {
		// Generate Deployment
		deployment := GenerateDeployment(obj, g.config, tsuboName)
		manifests.Deployments = append(manifests.Deployments, deployment)

		// Generate Service
		service := GenerateService(obj, g.config, tsuboName)
		manifests.Services = append(manifests.Services, service)
	}

	// Write manifests to files
	if err := g.writeManifests(namespaceManifest, manifests, tsuboName); err != nil {
		return nil, fmt.Errorf("failed to write manifests: %w", err)
	}

	return manifests, nil
}

// writeManifests writes all manifests to files
func (g *Generator) writeManifests(namespaceManifest string, manifests *ManifestSet, tsuboName string) error {
	// Create output directory
	outputDir := g.config.OutputDir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write namespace manifest
	namespacePath := filepath.Join(outputDir, "namespace.yaml")
	if err := os.WriteFile(namespacePath, []byte(namespaceManifest), 0644); err != nil {
		return fmt.Errorf("failed to write namespace manifest: %w", err)
	}

	// Write deployment manifests
	for i, deployment := range manifests.Deployments {
		filename := fmt.Sprintf("deployment-%d.yaml", i)
		path := filepath.Join(outputDir, filename)
		if err := os.WriteFile(path, []byte(deployment), 0644); err != nil {
			return fmt.Errorf("failed to write deployment manifest: %w", err)
		}
	}

	// Write service manifests
	for i, service := range manifests.Services {
		filename := fmt.Sprintf("service-%d.yaml", i)
		path := filepath.Join(outputDir, filename)
		if err := os.WriteFile(path, []byte(service), 0644); err != nil {
			return fmt.Errorf("failed to write service manifest: %w", err)
		}
	}

	fmt.Printf("✅ Generated Kubernetes manifests in: %s\n", outputDir)
	fmt.Printf("   - Namespace: %s\n", namespacePath)
	fmt.Printf("   - Deployments: %d\n", len(manifests.Deployments))
	fmt.Printf("   - Services: %d\n", len(manifests.Services))

	return nil
}
