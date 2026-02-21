package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	for _, obj := range tsuboDef.Objects {
		sm := generateServiceMonitor(obj, g.config)
		manifests.ServiceMonitors = append(manifests.ServiceMonitors, sm)

		filename := fmt.Sprintf("%s-service-monitor.yaml", obj.Name)
		path := filepath.Join(g.config.OutputDir, filename)
		if err := os.WriteFile(path, []byte(sm), 0644); err != nil {
			return nil, fmt.Errorf("failed to write ServiceMonitor for %s: %w", obj.Name, err)
		}
		fmt.Printf("  ✓ %s → %s\n", obj.Name, path)
	}

	return manifests, nil
}

// generateServiceMonitor generates a Prometheus Operator ServiceMonitor manifest
// from an ObjectRef. The generated manifest instructs Prometheus to scrape
// the /metrics endpoint of the service.
func generateServiceMonitor(obj types.ObjectRef, config *MonitorConfig) string {
	name := obj.Name
	namespace := config.Namespace
	interval := config.Interval

	// K8s port names must be lowercase alphanumeric + hyphens
	portName := strings.ToLower(strings.ReplaceAll(name, "_", "-"))

	var sb strings.Builder
	sb.WriteString("apiVersion: monitoring.coreos.com/v1\n")
	sb.WriteString("kind: ServiceMonitor\n")
	sb.WriteString("metadata:\n")
	sb.WriteString(fmt.Sprintf("  name: %s\n", name))
	sb.WriteString(fmt.Sprintf("  namespace: %s\n", namespace))
	sb.WriteString("  labels:\n")
	sb.WriteString(fmt.Sprintf("    app: %s\n", name))
	sb.WriteString("    managed-by: potter\n")
	sb.WriteString("spec:\n")
	sb.WriteString("  selector:\n")
	sb.WriteString("    matchLabels:\n")
	sb.WriteString(fmt.Sprintf("      app: %s\n", name))
	sb.WriteString("  endpoints:\n")
	sb.WriteString(fmt.Sprintf("    - port: %s\n", portName))
	sb.WriteString("      path: /metrics\n")
	sb.WriteString(fmt.Sprintf("      interval: %s\n", interval))
	sb.WriteString("  namespaceSelector:\n")
	sb.WriteString("    matchNames:\n")
	sb.WriteString(fmt.Sprintf("      - %s\n", namespace))

	return sb.String()
}
