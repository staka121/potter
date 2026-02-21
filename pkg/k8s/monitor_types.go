package k8s

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

// MonitorManifestSet contains all generated monitoring manifests
type MonitorManifestSet struct {
	Namespace       string
	ServiceMonitors []string
	PrometheusRules []string
}
