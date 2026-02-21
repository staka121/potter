package k8s

import (
	"fmt"
	"os"

	"github.com/staka121/potter/pkg/types"
)

// MonitorGenerator handles monitoring manifest generation
type MonitorGenerator struct {
	config *MonitorConfig
}

// NewMonitorGenerator creates a new monitor manifest generator
func NewMonitorGenerator(config *MonitorConfig) *MonitorGenerator {
	if config == nil {
		config = DefaultMonitorConfig()
	}
	return &MonitorGenerator{config: config}
}

// Generate generates all monitoring manifests from a TsuboDefinition
func (g *MonitorGenerator) Generate(tsuboDef *types.TsuboDefinition) (*MonitorManifestSet, error) {
	manifests := &MonitorManifestSet{
		Namespace:       g.config.Namespace,
		ServiceMonitors: make([]string, 0),
		PrometheusRules: make([]string, 0),
	}

	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("  Output directory: %s\n", g.config.OutputDir)
	fmt.Printf("  Objects found: %d\n", len(tsuboDef.Objects))
	for _, obj := range tsuboDef.Objects {
		fmt.Printf("    - %s\n", obj.Name)
	}

	return manifests, nil
}
