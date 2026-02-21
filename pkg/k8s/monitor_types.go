package k8s

import "github.com/staka121/potter/pkg/types"

// MonitorConfig contains configuration for monitoring manifest generation
type MonitorConfig struct {
	Namespace string
	OutputDir string
	Interval  string
}

// DefaultMonitorConfig returns default configuration
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		Namespace: "default",
		OutputDir: "monitor",
		Interval:  "15s",
	}
}

// MonitorTarget holds all data needed to generate monitoring manifests for a service
type MonitorTarget struct {
	Object      types.ObjectRef
	Performance *types.PerformanceConfig // nil if not defined in contract
}

// MonitorManifestSet contains all generated monitoring manifests
type MonitorManifestSet struct {
	Namespace       string
	ServiceMonitors []string
	PrometheusRules []string
}
