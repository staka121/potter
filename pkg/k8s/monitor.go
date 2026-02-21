package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

// Generate generates all monitoring manifests from a list of MonitorTargets.
// Each target provides service metadata and optional SLA performance config.
func (g *MonitorGenerator) Generate(targets []MonitorTarget) (*MonitorManifestSet, error) {
	manifests := &MonitorManifestSet{
		Namespace:       g.config.Namespace,
		ServiceMonitors: make([]string, 0),
		PrometheusRules: make([]string, 0),
	}

	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, target := range targets {
		// ServiceMonitor
		sm := generateServiceMonitor(target.Object, g.config)
		manifests.ServiceMonitors = append(manifests.ServiceMonitors, sm)

		smPath := filepath.Join(g.config.OutputDir, fmt.Sprintf("%s-service-monitor.yaml", target.Object.Name))
		if err := os.WriteFile(smPath, []byte(sm), 0644); err != nil {
			return nil, fmt.Errorf("failed to write ServiceMonitor for %s: %w", target.Object.Name, err)
		}
		fmt.Printf("  ✓ ServiceMonitor  %s → %s\n", target.Object.Name, smPath)

		// PrometheusRule (only when performance.latency is defined)
		if target.Performance != nil && latencyHasAny(target.Performance.Latency) {
			pr, err := generatePrometheusRule(target.Object.Name, target.Performance.Latency, g.config)
			if err != nil {
				return nil, fmt.Errorf("failed to generate PrometheusRule for %s: %w", target.Object.Name, err)
			}
			manifests.PrometheusRules = append(manifests.PrometheusRules, pr)

			prPath := filepath.Join(g.config.OutputDir, fmt.Sprintf("%s-prometheus-rule.yaml", target.Object.Name))
			if err := os.WriteFile(prPath, []byte(pr), 0644); err != nil {
				return nil, fmt.Errorf("failed to write PrometheusRule for %s: %w", target.Object.Name, err)
			}
			fmt.Printf("  ✓ PrometheusRule  %s → %s\n", target.Object.Name, prPath)
		}
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

// generatePrometheusRule generates a Prometheus Operator PrometheusRule manifest
// from a service name and its SLA latency thresholds.
func generatePrometheusRule(name string, latency types.LatencyConfig, config *MonitorConfig) (string, error) {
	namespace := config.Namespace

	type rule struct {
		quantile float64
		label    string
		threshold float64
		severity string
	}

	var rules []rule

	for _, entry := range []struct {
		raw      string
		quantile float64
		label    string
		severity string
	}{
		{latency.P50, 0.50, "p50", "warning"},
		{latency.P95, 0.95, "p95", "warning"},
		{latency.P99, 0.99, "p99", "critical"},
	} {
		if entry.raw == "" {
			continue
		}
		secs, err := parseDurationToSeconds(entry.raw)
		if err != nil {
			return "", fmt.Errorf("invalid latency value %q: %w", entry.raw, err)
		}
		rules = append(rules, rule{entry.quantile, entry.label, secs, entry.severity})
	}

	var sb strings.Builder
	sb.WriteString("apiVersion: monitoring.coreos.com/v1\n")
	sb.WriteString("kind: PrometheusRule\n")
	sb.WriteString("metadata:\n")
	sb.WriteString(fmt.Sprintf("  name: %s-sla\n", name))
	sb.WriteString(fmt.Sprintf("  namespace: %s\n", namespace))
	sb.WriteString("  labels:\n")
	sb.WriteString(fmt.Sprintf("    app: %s\n", name))
	sb.WriteString("    managed-by: potter\n")
	sb.WriteString("spec:\n")
	sb.WriteString("  groups:\n")
	sb.WriteString(fmt.Sprintf("    - name: %s.sla\n", name))
	sb.WriteString("      rules:\n")

	for _, r := range rules {
		alertName := fmt.Sprintf("SLAViolation%s", strings.ToUpper(r.label))
		expr := fmt.Sprintf(
			`histogram_quantile(%.2f, rate(http_request_duration_seconds_bucket{app="%s"}[5m])) > %s`,
			r.quantile, name, formatSeconds(r.threshold),
		)
		sb.WriteString(fmt.Sprintf("        - alert: %s\n", alertName))
		sb.WriteString(fmt.Sprintf("          expr: %s\n", expr))
		sb.WriteString("          for: 1m\n")
		sb.WriteString("          labels:\n")
		sb.WriteString(fmt.Sprintf("            severity: %s\n", r.severity))
		sb.WriteString(fmt.Sprintf("            service: %s\n", name))
		sb.WriteString("          annotations:\n")
		sb.WriteString(fmt.Sprintf("            summary: \"%s latency SLA violated for %s\"\n", strings.ToUpper(r.label), name))
		sb.WriteString(fmt.Sprintf("            description: \"%s latency exceeds %s\"\n", strings.ToUpper(r.label), latencyRaw(r.label, latency)))
	}

	return sb.String(), nil
}

// parseDurationToSeconds parses a duration string like "50ms", "1s", "200ms" into seconds.
func parseDurationToSeconds(s string) (float64, error) {
	s = strings.TrimSpace(s)
	switch {
	case strings.HasSuffix(s, "ms"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "ms"), 64)
		return v / 1000, err
	case strings.HasSuffix(s, "us"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "us"), 64)
		return v / 1_000_000, err
	case strings.HasSuffix(s, "s"):
		return strconv.ParseFloat(strings.TrimSuffix(s, "s"), 64)
	default:
		return 0, fmt.Errorf("unsupported duration unit in %q (expected ms, us, or s)", s)
	}
}

// formatSeconds formats a float64 seconds value without trailing zeros.
func formatSeconds(v float64) string {
	s := strconv.FormatFloat(v, 'f', 6, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// latencyRaw returns the raw string for a given percentile label.
func latencyRaw(label string, l types.LatencyConfig) string {
	switch label {
	case "p50":
		return l.P50
	case "p95":
		return l.P95
	case "p99":
		return l.P99
	}
	return ""
}

// latencyHasAny returns true if at least one latency threshold is defined.
func latencyHasAny(l types.LatencyConfig) bool {
	return l.P50 != "" || l.P95 != "" || l.P99 != ""
}
