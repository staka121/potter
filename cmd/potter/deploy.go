package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/staka121/potter/internal/parser"
	"github.com/staka121/potter/pkg/k8s"
)

func runDeploy(args []string) error {
	if len(args) == 0 {
		printDeployUsage()
		return nil
	}

	subcommand := args[0]

	switch subcommand {
	case "generate":
		return runDeployGenerate(args[1:])
	case "help", "--help", "-h":
		printDeployUsage()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown deploy subcommand: %s\n\n", subcommand)
		printDeployUsage()
		return fmt.Errorf("unknown deploy subcommand: %s", subcommand)
	}
}

func runDeployGenerate(args []string) error {
	fs := flag.NewFlagSet("deploy generate", flag.ExitOnError)

	namespace := fs.String("namespace", "default", "Kubernetes namespace")
	outputDir := fs.String("output", "k8s", "Output directory for manifests")
	registry := fs.String("registry", "", "Docker image registry (e.g., docker.io/myorg)")
	imageTag := fs.String("tag", "latest", "Docker image tag")
	replicas := fs.Int("replicas", 1, "Default number of replicas")
	ingressEnabled := fs.Bool("ingress", true, "Enable Ingress generation (replaces gateway-service)")
	ingressHost := fs.String("ingress-host", "", "Ingress host (e.g., todo.example.com)")
	ingressClass := fs.String("ingress-class", "nginx", "Ingress class name")
	helpFlag := fs.Bool("help", false, "Show help for deploy generate command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printDeployGenerateUsage()
		return nil
	}

	// Get tsubo file path
	args = fs.Args()
	if len(args) == 0 {
		return fmt.Errorf("tsubo file path required. Usage: potter deploy generate <tsubo-file> [options]")
	}

	tsuboFile := args[0]

	// Verify file exists
	if _, err := os.Stat(tsuboFile); os.IsNotExist(err) {
		return fmt.Errorf("tsubo file not found: %s", tsuboFile)
	}

	printDeployGenerateHeader()

	// Step 1: Parse tsubo file
	fmt.Printf("%s[Step 1] Parsing Tsubo definition%s\n", colorYellow, colorReset)
	tsuboDef, err := parser.ParseTsuboFile(tsuboFile)
	if err != nil {
		return fmt.Errorf("failed to parse tsubo file: %w", err)
	}

	fmt.Printf("  %sTsubo: %s%s\n", colorGreen, tsuboDef.Tsubo.Name, colorReset)
	fmt.Printf("  Objects: %d\n", len(tsuboDef.Objects))
	fmt.Println()

	// Step 2: Generate Kubernetes manifests
	fmt.Printf("%s[Step 2] Generating Kubernetes manifests%s\n", colorYellow, colorReset)

	// Configure Ingress
	ingressConfig := &k8s.IngressConfig{
		Enabled:      *ingressEnabled,
		Host:         *ingressHost,
		TLSEnabled:   false,
		IngressClass: *ingressClass,
		Annotations:  make(map[string]string),
	}

	config := &k8s.GeneratorConfig{
		Namespace:       *namespace,
		OutputDir:       *outputDir,
		ImageRegistry:   *registry,
		ImageTag:        *imageTag,
		DefaultReplicas: int32(*replicas),
		Ingress:         ingressConfig,
	}

	generator := k8s.NewGenerator(config)
	manifests, err := generator.Generate(tsuboDef)
	if err != nil {
		return fmt.Errorf("failed to generate manifests: %w", err)
	}

	fmt.Println()
	fmt.Printf("%s✅ Kubernetes manifests generated successfully!%s\n", colorGreen, colorReset)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Review manifests in: %s\n", *outputDir)
	fmt.Printf("  2. Apply to cluster: kubectl apply -f %s/\n", *outputDir)
	fmt.Printf("  3. Check status: kubectl get pods -n %s\n", manifests.Namespace)

	return nil
}

func printDeployUsage() {
	fmt.Println("Potter Deploy - Kubernetes Deployment Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  potter deploy <subcommand> [options]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  generate   Generate Kubernetes manifests from tsubo definition")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter deploy generate app.tsubo.yaml")
	fmt.Println("  potter deploy generate --namespace prod --output k8s-prod app.tsubo.yaml")
	fmt.Println("  potter deploy generate --registry docker.io/myorg --tag v1.0.0 app.tsubo.yaml")
	fmt.Println()
}

func printDeployGenerateUsage() {
	fmt.Println("Potter Deploy Generate - Generate Kubernetes Manifests")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  potter deploy generate [options] <tsubo-file>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --namespace string      Kubernetes namespace (default: default)")
	fmt.Println("  --output string         Output directory for manifests (default: k8s)")
	fmt.Println("  --registry string       Docker image registry (e.g., docker.io/myorg)")
	fmt.Println("  --tag string            Docker image tag (default: latest)")
	fmt.Println("  --replicas int          Default number of replicas (default: 1)")
	fmt.Println("  --ingress               Enable Ingress generation (default: true)")
	fmt.Println("  --ingress-host string   Ingress host (e.g., todo.example.com)")
	fmt.Println("  --ingress-class string  Ingress class name (default: nginx)")
	fmt.Println("  --help                  Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter deploy generate app.tsubo.yaml")
	fmt.Println("  potter deploy generate --namespace production app.tsubo.yaml")
	fmt.Println("  potter deploy generate --registry gcr.io/myproject --tag v1.0.0 app.tsubo.yaml")
	fmt.Println("  potter deploy generate --replicas 3 --output ./manifests app.tsubo.yaml")
	fmt.Println("  potter deploy generate --ingress-host todo.example.com app.tsubo.yaml")
	fmt.Println()
}

func printDeployGenerateHeader() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  Potter Deploy - Kubernetes Manifest Generator")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
}
