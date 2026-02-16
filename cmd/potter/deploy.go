package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	case "apply":
		return runDeployApply(args[1:])
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
	fmt.Println("  apply      Apply manifests to Kubernetes cluster")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter deploy generate app.tsubo.yaml")
	fmt.Println("  potter deploy apply")
	fmt.Println("  potter deploy apply --namespace production")
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

func runDeployApply(args []string) error {
	fs := flag.NewFlagSet("deploy apply", flag.ExitOnError)

	manifestDir := fs.String("manifests", "k8s", "Directory containing K8s manifests")
	namespace := fs.String("namespace", "", "Kubernetes namespace (overrides manifest namespace)")
	wait := fs.Bool("wait", true, "Wait for rollout to complete")
	timeout := fs.Duration("timeout", 5*time.Minute, "Timeout for rollout")
	helpFlag := fs.Bool("help", false, "Show help for deploy apply command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printDeployApplyUsage()
		return nil
	}

	printDeployApplyHeader()

	// Step 1: Check if kubectl is available
	fmt.Printf("%s[Step 1] Checking kubectl availability%s\n", colorYellow, colorReset)
	if err := checkKubectlAvailable(); err != nil {
		return fmt.Errorf("kubectl not found: %w\nPlease install kubectl: https://kubernetes.io/docs/tasks/tools/", err)
	}
	fmt.Printf("  %s✓ kubectl is available%s\n", colorGreen, colorReset)
	fmt.Println()

	// Step 2: Check if manifest directory exists
	fmt.Printf("%s[Step 2] Checking manifest directory%s\n", colorYellow, colorReset)
	if _, err := os.Stat(*manifestDir); os.IsNotExist(err) {
		return fmt.Errorf("manifest directory not found: %s\nRun 'potter deploy generate' first", *manifestDir)
	}

	// Count manifest files
	manifestFiles, err := filepath.Glob(filepath.Join(*manifestDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to list manifests: %w", err)
	}
	if len(manifestFiles) == 0 {
		return fmt.Errorf("no manifest files found in: %s", *manifestDir)
	}

	fmt.Printf("  %sFound %d manifest file(s)%s\n", colorGreen, len(manifestFiles), colorReset)
	fmt.Println()

	// Step 3: Apply manifests
	fmt.Printf("%s[Step 3] Applying manifests to cluster%s\n", colorYellow, colorReset)

	applyArgs := []string{"apply", "-f", *manifestDir}
	if *namespace != "" {
		applyArgs = append(applyArgs, "-n", *namespace)
	}

	cmd := exec.Command("kubectl", applyArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}
	fmt.Println()

	// Step 4: Wait for rollout if requested
	if *wait {
		fmt.Printf("%s[Step 4] Waiting for rollout to complete%s\n", colorYellow, colorReset)

		// Get namespace from manifests if not specified
		ns := *namespace
		if ns == "" {
			ns = getNamespaceFromManifests(*manifestDir)
		}

		if err := waitForRollout(ns, *timeout); err != nil {
			fmt.Printf("%s⚠ Rollout monitoring failed: %v%s\n", colorYellow, err, colorReset)
			fmt.Println("Resources have been applied, but rollout status could not be verified.")
		} else {
			fmt.Printf("  %s✓ Rollout completed successfully%s\n", colorGreen, colorReset)
		}
		fmt.Println()
	}

	// Step 5: Show status
	fmt.Printf("%s[Step 5] Deployment status%s\n", colorYellow, colorReset)
	showDeploymentStatus(*namespace)

	fmt.Println()
	fmt.Printf("%s✅ Deployment completed successfully!%s\n", colorGreen, colorReset)
	fmt.Println()

	return nil
}

func checkKubectlAvailable() error {
	cmd := exec.Command("kubectl", "version", "--client", "--short")
	return cmd.Run()
}

func getNamespaceFromManifests(manifestDir string) string {
	// Try to read namespace from namespace.yaml
	namespacePath := filepath.Join(manifestDir, "namespace.yaml")
	data, err := os.ReadFile(namespacePath)
	if err != nil {
		return "default"
	}

	// Simple parsing - look for "name:" after "metadata:"
	lines := strings.Split(string(data), "\n")
	inMetadata := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "metadata:") {
			inMetadata = true
			continue
		}
		if inMetadata && strings.HasPrefix(trimmed, "name:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return "default"
}

func waitForRollout(namespace string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Check deployments
		cmd := exec.Command("kubectl", "get", "deployments", "-n", namespace, "-o", "jsonpath={.items[*].metadata.name}")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get deployments: %w", err)
		}

		deployments := strings.Fields(string(output))
		if len(deployments) == 0 {
			// No deployments, consider it successful
			return nil
		}

		allReady := true
		for _, deployment := range deployments {
			cmd := exec.Command("kubectl", "rollout", "status", "deployment/"+deployment, "-n", namespace, "--timeout=10s")
			if err := cmd.Run(); err != nil {
				allReady = false
				break
			}
		}

		if allReady {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for rollout")
}

func showDeploymentStatus(namespace string) {
	if namespace == "" {
		namespace = "default"
	}

	fmt.Printf("  Pods:\n")
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	fmt.Println()
	fmt.Printf("  Services:\n")
	cmd = exec.Command("kubectl", "get", "svc", "-n", namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func printDeployApplyUsage() {
	fmt.Println("Potter Deploy Apply - Apply Manifests to Kubernetes")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  potter deploy apply [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --manifests string   Directory containing K8s manifests (default: k8s)")
	fmt.Println("  --namespace string   Kubernetes namespace (overrides manifest namespace)")
	fmt.Println("  --wait               Wait for rollout to complete (default: true)")
	fmt.Println("  --timeout duration   Timeout for rollout (default: 5m)")
	fmt.Println("  --help               Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter deploy apply")
	fmt.Println("  potter deploy apply --manifests k8s-prod")
	fmt.Println("  potter deploy apply --namespace production")
	fmt.Println("  potter deploy apply --wait=false")
	fmt.Println("  potter deploy apply --timeout 10m")
	fmt.Println()
}

func printDeployApplyHeader() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  Potter Deploy - Kubernetes Apply")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
}
