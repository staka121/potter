package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/staka121/potter/internal/parser"
	"github.com/staka121/potter/pkg/dashboard"
	"github.com/staka121/potter/pkg/k8s"
	prometheusclient "github.com/staka121/potter/pkg/prometheus"
)

func runMonitor(args []string) error {
	if len(args) == 0 {
		printMonitorUsage()
		return nil
	}

	subcommand := args[0]

	switch subcommand {
	case "generate":
		return runMonitorGenerate(args[1:])
	case "apply":
		return runMonitorApply(args[1:])
	case "dashboard":
		return runMonitorDashboard(args[1:])
	case "help", "--help", "-h":
		printMonitorUsage()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown monitor subcommand: %s\n\n", subcommand)
		printMonitorUsage()
		return fmt.Errorf("unknown monitor subcommand: %s", subcommand)
	}
}

func runMonitorGenerate(args []string) error {
	fs := flag.NewFlagSet("monitor generate", flag.ExitOnError)

	namespace := fs.String("namespace", "default", "Kubernetes namespace")
	interval := fs.String("interval", "15s", "Scrape interval for ServiceMonitor")
	helpFlag := fs.Bool("help", false, "Show help for monitor generate command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printMonitorGenerateUsage()
		return nil
	}

	args = fs.Args()
	if len(args) == 0 {
		return fmt.Errorf("tsubo file path required. Usage: potter monitor generate <tsubo-file> [options]")
	}

	tsuboFile := args[0]

	if _, err := os.Stat(tsuboFile); os.IsNotExist(err) {
		return fmt.Errorf("tsubo file not found: %s", tsuboFile)
	}

	printMonitorGenerateHeader()

	// Step 1: Parse tsubo file
	fmt.Printf("%s[Step 1] Parsing Tsubo definition%s\n", colorYellow, colorReset)
	tsuboDef, err := parser.ParseTsuboFile(tsuboFile)
	if err != nil {
		return fmt.Errorf("failed to parse tsubo file: %w", err)
	}

	fmt.Printf("  %sTsubo: %s%s\n", colorGreen, tsuboDef.Tsubo.Name, colorReset)
	fmt.Printf("  Objects: %d\n", len(tsuboDef.Objects))
	fmt.Println()

	// Step 2: Load contracts and build monitor targets
	fmt.Printf("%s[Step 2] Loading contracts%s\n", colorYellow, colorReset)

	tsuboDir := filepath.Dir(tsuboFile)
	targets := make([]k8s.MonitorTarget, 0, len(tsuboDef.Objects))

	for _, obj := range tsuboDef.Objects {
		target := k8s.MonitorTarget{Object: obj}

		if obj.Contract != "" {
			contractPath := filepath.Join(tsuboDir, obj.Contract)
			objDef, err := parser.ParseObjectFile(contractPath)
			if err != nil {
				fmt.Printf("  %s⚠ %s: contract not found, skipping SLA rules%s\n", colorYellow, obj.Name, colorReset)
			} else if objDef.Performance.Latency.P50 != "" || objDef.Performance.Latency.P95 != "" || objDef.Performance.Latency.P99 != "" {
				target.Performance = &objDef.Performance
				fmt.Printf("  %s✓ %s: SLA defined (p50=%s p95=%s p99=%s)%s\n",
					colorGreen, obj.Name,
					objDef.Performance.Latency.P50,
					objDef.Performance.Latency.P95,
					objDef.Performance.Latency.P99,
					colorReset,
				)
			} else {
				fmt.Printf("  - %s: no performance SLA defined\n", obj.Name)
			}
		}

		targets = append(targets, target)
	}
	fmt.Println()

	// Step 3: Generate monitoring manifests
	fmt.Printf("%s[Step 3] Generating monitoring manifests%s\n", colorYellow, colorReset)

	outputDir := filepath.Join(tsuboDir, "monitor")
	config := &k8s.MonitorConfig{
		Namespace: *namespace,
		OutputDir: outputDir,
		Interval:  *interval,
	}

	generator := k8s.NewMonitorGenerator(config)
	manifests, err := generator.Generate(targets)
	if err != nil {
		return fmt.Errorf("failed to generate manifests: %w", err)
	}

	fmt.Println()
	fmt.Printf("%s✅ Monitoring manifests generated successfully!%s\n", colorGreen, colorReset)
	fmt.Printf("   - ServiceMonitors: %d\n", len(manifests.ServiceMonitors))
	fmt.Printf("   - PrometheusRules: %d\n", len(manifests.PrometheusRules))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Review manifests in: %s\n", outputDir)
	fmt.Printf("  2. Apply to cluster:    potter monitor apply %s\n", tsuboFile)
	fmt.Printf("  3. Namespace:           %s\n", manifests.Namespace)

	return nil
}

func runMonitorApply(args []string) error {
	fs := flag.NewFlagSet("monitor apply", flag.ExitOnError)

	manifestDir := fs.String("manifests", "", "Directory containing monitoring manifests (default: <tsubo-dir>/monitor)")
	helpFlag := fs.Bool("help", false, "Show help for monitor apply command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printMonitorApplyUsage()
		return nil
	}

	args = fs.Args()
	if len(args) == 0 && *manifestDir == "" {
		return fmt.Errorf("tsubo file path required. Usage: potter monitor apply <tsubo-file> [options]")
	}

	// Resolve manifest directory
	dir := *manifestDir
	if dir == "" {
		tsuboFile := args[0]
		if _, err := os.Stat(tsuboFile); os.IsNotExist(err) {
			return fmt.Errorf("tsubo file not found: %s", tsuboFile)
		}
		dir = filepath.Join(filepath.Dir(tsuboFile), "monitor")
	}

	printMonitorApplyHeader()

	// Step 1: Check kubectl
	fmt.Printf("%s[Step 1] Checking kubectl availability%s\n", colorYellow, colorReset)
	if err := exec.Command("kubectl", "version", "--client", "--short").Run(); err != nil {
		return fmt.Errorf("kubectl not found: %w\nPlease install kubectl: https://kubernetes.io/docs/tasks/tools/", err)
	}
	fmt.Printf("  %s✓ kubectl is available%s\n", colorGreen, colorReset)
	fmt.Println()

	// Step 2: Check Prometheus Operator CRDs
	fmt.Printf("%s[Step 2] Checking Prometheus Operator%s\n", colorYellow, colorReset)
	if err := checkPrometheusOperator(); err != nil {
		return err
	}
	fmt.Printf("  %s✓ Prometheus Operator CRDs found%s\n", colorGreen, colorReset)
	fmt.Println()

	// Step 3: Check manifest directory
	fmt.Printf("%s[Step 3] Checking manifest directory%s\n", colorYellow, colorReset)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("manifest directory not found: %s\nRun 'potter monitor generate' first", dir)
	}

	manifestFiles, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to list manifests: %w", err)
	}
	if len(manifestFiles) == 0 {
		return fmt.Errorf("no manifest files found in: %s", dir)
	}

	fmt.Printf("  %sFound %d manifest file(s)%s\n", colorGreen, len(manifestFiles), colorReset)
	fmt.Println()

	// Step 4: Apply manifests
	fmt.Printf("%s[Step 4] Applying monitoring manifests to cluster%s\n", colorYellow, colorReset)

	applyCmd := exec.Command("kubectl", "apply", "-f", dir)
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr

	if err := applyCmd.Run(); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}
	fmt.Println()

	// Step 5: Show status
	fmt.Printf("%s[Step 5] Applied resources%s\n", colorYellow, colorReset)
	showMonitorStatus(dir)

	fmt.Println()
	fmt.Printf("%s✅ Monitoring manifests applied successfully!%s\n", colorGreen, colorReset)
	fmt.Println()

	return nil
}

// checkPrometheusOperator verifies that the required Prometheus Operator CRDs are installed.
func checkPrometheusOperator() error {
	crds := []string{
		"servicemonitors.monitoring.coreos.com",
		"prometheusrules.monitoring.coreos.com",
	}
	for _, crd := range crds {
		if err := exec.Command("kubectl", "get", "crd", crd).Run(); err != nil {
			return fmt.Errorf(
				"Prometheus Operator CRD not found: %s\n"+
					"  Please install Prometheus Operator first:\n"+
					"  https://prometheus-operator.dev/docs/getting-started/installation/",
				crd,
			)
		}
	}
	return nil
}

// showMonitorStatus lists ServiceMonitors and PrometheusRules in the manifest directory's namespace.
func showMonitorStatus(manifestDir string) {
	// Detect namespace from manifest files
	ns := detectNamespaceFromDir(manifestDir)

	fmt.Printf("  ServiceMonitors (namespace: %s):\n", ns)
	smCmd := exec.Command("kubectl", "get", "servicemonitors", "-n", ns)
	smCmd.Stdout = os.Stdout
	smCmd.Stderr = os.Stderr
	smCmd.Run()

	fmt.Println()
	fmt.Printf("  PrometheusRules (namespace: %s):\n", ns)
	prCmd := exec.Command("kubectl", "get", "prometheusrules", "-n", ns)
	prCmd.Stdout = os.Stdout
	prCmd.Stderr = os.Stderr
	prCmd.Run()
}

// detectNamespaceFromDir extracts the namespace from the first YAML file in the directory.
func detectNamespaceFromDir(dir string) string {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil || len(files) == 0 {
		return "default"
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		return "default"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "namespace:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "default"
}

func printMonitorUsage() {
	fmt.Println("Potter Monitor - Contract-driven Monitoring for Kubernetes")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  potter monitor <subcommand> [options]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  generate    Generate ServiceMonitor / PrometheusRule manifests from Contract")
	fmt.Println("  apply       Apply monitoring manifests to Kubernetes cluster")
	fmt.Println("  dashboard   Real-time SLA dashboard via Prometheus API")
	fmt.Println("  help        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter monitor generate app.tsubo.yaml")
	fmt.Println("  potter monitor apply app.tsubo.yaml")
	fmt.Println("  potter monitor dashboard app.tsubo.yaml")
	fmt.Println()
}

func printMonitorGenerateUsage() {
	fmt.Println("Potter Monitor Generate - Generate Monitoring Manifests from Contract")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  potter monitor generate [options] <tsubo-file>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --namespace string   Kubernetes namespace (default: default)")
	fmt.Println("  --interval string    Scrape interval for ServiceMonitor (default: 15s)")
	fmt.Println("  --help               Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter monitor generate app.tsubo.yaml")
	fmt.Println("  potter monitor generate --namespace monitoring app.tsubo.yaml")
	fmt.Println("  potter monitor generate --interval 30s app.tsubo.yaml")
	fmt.Println()
}

func printMonitorApplyUsage() {
	fmt.Println("Potter Monitor Apply - Apply Monitoring Manifests to Kubernetes")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  potter monitor apply [options] <tsubo-file>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --manifests string   Directory containing monitoring manifests (default: <tsubo-dir>/monitor)")
	fmt.Println("  --help               Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter monitor apply app.tsubo.yaml")
	fmt.Println("  potter monitor apply --manifests ./monitor app.tsubo.yaml")
	fmt.Println()
}

func printMonitorGenerateHeader() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  Potter Monitor - Monitoring Manifest Generator")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
}

func printMonitorApplyHeader() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  Potter Monitor - Apply to Kubernetes")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
}

func runMonitorDashboard(args []string) error {
	fs := flag.NewFlagSet("monitor dashboard", flag.ExitOnError)

	prometheusEndpoint := fs.String("prometheus", "http://localhost:9090", "Prometheus API endpoint")
	interval := fs.Duration("interval", 5*time.Second, "Dashboard refresh interval")
	helpFlag := fs.Bool("help", false, "Show help for monitor dashboard command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printMonitorDashboardUsage()
		return nil
	}

	args = fs.Args()
	if len(args) == 0 {
		return fmt.Errorf("tsubo file path required. Usage: potter monitor dashboard <tsubo-file> [options]")
	}

	tsuboFile := args[0]
	if _, err := os.Stat(tsuboFile); os.IsNotExist(err) {
		return fmt.Errorf("tsubo file not found: %s", tsuboFile)
	}

	// Parse tsubo and load contracts to build MonitorTargets
	tsuboDef, err := parser.ParseTsuboFile(tsuboFile)
	if err != nil {
		return fmt.Errorf("failed to parse tsubo file: %w", err)
	}

	tsuboDir := filepath.Dir(tsuboFile)
	targets := make([]k8s.MonitorTarget, 0, len(tsuboDef.Objects))
	for _, obj := range tsuboDef.Objects {
		target := k8s.MonitorTarget{Object: obj}
		if obj.Contract != "" {
			contractPath := filepath.Join(tsuboDir, obj.Contract)
			if objDef, err := parser.ParseObjectFile(contractPath); err == nil {
				if objDef.Performance.Latency.P50 != "" || objDef.Performance.Latency.P95 != "" || objDef.Performance.Latency.P99 != "" {
					target.Performance = &objDef.Performance
				}
			}
		}
		targets = append(targets, target)
	}

	client := prometheusclient.NewClient(*prometheusEndpoint)

	// Handle Ctrl+C gracefully
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("Starting dashboard (press Ctrl+C to exit)...\n")
	time.Sleep(500 * time.Millisecond)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// Initial render
	render(tsuboDef.Tsubo.Name, *prometheusEndpoint, targets, client)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nExiting dashboard.")
			return nil
		case <-ticker.C:
			render(tsuboDef.Tsubo.Name, *prometheusEndpoint, targets, client)
		}
	}
}

// render fetches metrics for all targets and draws the dashboard.
func render(tsuboName, prometheusEndpoint string, targets []k8s.MonitorTarget, client *prometheusclient.Client) {
	rows := make([]dashboard.Row, 0, len(targets))
	for _, t := range targets {
		p50, p95, p99, err := client.QueryLatency(t.Object.Name)
		row := dashboard.Row{
			ServiceName: t.Object.Name,
			P50Ms:       p50,
			P95Ms:       p95,
			P99Ms:       p99,
		}
		if err != nil {
			row.P50Ms, row.P95Ms, row.P99Ms = -1, -1, -1
		}
		if t.Performance != nil {
			row.SLA = &t.Performance.Latency
		}
		rows = append(rows, row)
	}
	dashboard.Render(tsuboName, prometheusEndpoint, rows, time.Now())
}

func printMonitorDashboardUsage() {
	fmt.Println("Potter Monitor Dashboard - Real-time SLA Dashboard")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  potter monitor dashboard [options] <tsubo-file>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --prometheus string   Prometheus API endpoint (default: http://localhost:9090)")
	fmt.Println("  --interval duration   Refresh interval (default: 5s)")
	fmt.Println("  --help                Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter monitor dashboard app.tsubo.yaml")
	fmt.Println("  potter monitor dashboard --prometheus http://prometheus.example.com app.tsubo.yaml")
	fmt.Println("  potter monitor dashboard --interval 10s app.tsubo.yaml")
	fmt.Println()
	fmt.Println("Tip (Kubernetes):")
	fmt.Println("  kubectl port-forward svc/prometheus 9090:9090 -n monitoring")
	fmt.Println("  potter monitor dashboard app.tsubo.yaml")
	fmt.Println()
}
