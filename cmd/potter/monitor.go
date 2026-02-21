package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/staka121/potter/internal/parser"
	"github.com/staka121/potter/pkg/k8s"
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

	// Step 2: Check manifest directory
	fmt.Printf("%s[Step 2] Checking manifest directory%s\n", colorYellow, colorReset)
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

	// Step 3: Apply manifests
	fmt.Printf("%s[Step 3] Applying monitoring manifests to cluster%s\n", colorYellow, colorReset)

	cmd := exec.Command("kubectl", "apply", "-f", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	fmt.Println()
	fmt.Printf("%s✅ Monitoring manifests applied successfully!%s\n", colorGreen, colorReset)
	fmt.Println()

	return nil
}

func printMonitorUsage() {
	fmt.Println("Potter Monitor - Contract-driven Monitoring for Kubernetes")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  potter monitor <subcommand> [options]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  generate   Generate ServiceMonitor / PrometheusRule manifests from Contract")
	fmt.Println("  apply      Apply monitoring manifests to Kubernetes cluster")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter monitor generate app.tsubo.yaml")
	fmt.Println("  potter monitor apply app.tsubo.yaml")
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
