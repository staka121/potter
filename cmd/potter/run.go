package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/staka121/potter/internal/parser"
)

func runRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	helpFlag := fs.Bool("help", false, "Show help for run command")
	serviceFlag := fs.String("service", "", "Run specific service only")
	detachFlag := fs.Bool("d", false, "Run in detached mode")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printRunUsage()
		return nil
	}

	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sPotter Run%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Println()

	// Get tsubo file path (required)
	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		return fmt.Errorf("tsubo file path required. Usage: potter run [options] <tsubo-file>")
	}

	tsuboFile := remainingArgs[0]
	tsuboDir := filepath.Dir(tsuboFile)
	implDir := filepath.Join(tsuboDir, "implementations")

	if _, err := os.Stat(implDir); os.IsNotExist(err) {
		return fmt.Errorf("implementations directory not found: %s\nRun 'potter build %s' first to generate implementations", implDir, tsuboFile)
	}

	// Parse tsubo file to get network and startup order
	tsuboDef, err := parser.ParseTsuboFile(tsuboFile)
	if err != nil {
		return fmt.Errorf("failed to parse tsubo file: %w", err)
	}

	// Create Docker network if it doesn't exist
	// Using the network name defined in tsubo.yaml deployment.network.name
	networkName := "tsubo-network"
	fmt.Printf("Ensuring Docker network '%s' exists...\n", networkName)

	// Check if network exists
	checkCmd := exec.Command("docker", "network", "inspect", networkName)
	if err := checkCmd.Run(); err != nil {
		// Network doesn't exist, create it
		createCmd := exec.Command("docker", "network", "create", networkName)
		createCmd.Stdout = os.Stdout
		createCmd.Stderr = os.Stderr
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("failed to create network %s: %w", networkName, err)
		}
		fmt.Printf("  %s✓ Network created%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("  %s✓ Network already exists%s\n", colorGreen, colorReset)
	}
	fmt.Println()

	// Build service list based on dependencies
	// First, services without dependencies, then services with dependencies
	var servicesNoDeps []string
	var servicesWithDeps []string

	// Create a map for quick lookup
	serviceMap := make(map[string]bool)
	entries, err := os.ReadDir(implDir)
	if err != nil {
		return fmt.Errorf("failed to read implementations directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			serviceMap[entry.Name()] = true
		}
	}

	// Separate services by dependencies
	for _, objRef := range tsuboDef.Objects {
		// Skip if not in implementations directory
		if !serviceMap[objRef.Name] {
			continue
		}
		// Filter by service name if specified
		if *serviceFlag != "" && objRef.Name != *serviceFlag {
			continue
		}

		if len(objRef.Dependencies) == 0 {
			servicesNoDeps = append(servicesNoDeps, objRef.Name)
		} else {
			servicesWithDeps = append(servicesWithDeps, objRef.Name)
		}
	}

	// Combine: services without dependencies first, then services with dependencies
	services := append(servicesNoDeps, servicesWithDeps...)

	if len(services) == 0 {
		if *serviceFlag != "" {
			return fmt.Errorf("service not found: %s", *serviceFlag)
		}
		return fmt.Errorf("no services found in %s", implDir)
	}

	fmt.Printf("Starting %d service(s)...\n\n", len(services))

	// Always start services in detached mode first
	for _, service := range services {
		fmt.Printf("%s[%s]%s\n", colorYellow, service, colorReset)

		serviceDir := filepath.Join(implDir, service)

		// Check for docker-compose.yml
		composeFile := filepath.Join(serviceDir, "docker-compose.yml")
		if _, err := os.Stat(composeFile); os.IsNotExist(err) {
			fmt.Printf("  %s⚠ No docker-compose.yml found%s\n", colorYellow, colorReset)
			fmt.Println()
			continue
		}

		// Always run docker-compose up -d to start services in background
		cmdArgs := []string{"compose", "up", "-d"}

		cmd := exec.Command("docker", cmdArgs...)
		cmd.Dir = serviceDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("  %s✗ Failed to start service%s\n", colorRed, colorReset)
			return fmt.Errorf("failed to start %s: %w", service, err)
		}

		fmt.Printf("  %s✓ Service started%s\n", colorGreen, colorReset)
		fmt.Println()
	}

	// Auto-start gateway-service if it exists (implicit API Gateway for Tsubo encapsulation)
	// Gateway is started last, after all other services are up
	if *serviceFlag == "" {
		gatewayDir := filepath.Join(implDir, "gateway-service")
		gatewayComposeFile := filepath.Join(gatewayDir, "docker-compose.yml")

		if _, err := os.Stat(gatewayComposeFile); err == nil {
			fmt.Printf("%s[gateway-service]%s (auto-generated API Gateway)\n", colorYellow, colorReset)

			cmd := exec.Command("docker", "compose", "up", "-d")
			cmd.Dir = gatewayDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Printf("  %s✗ Failed to start gateway-service%s\n", colorRed, colorReset)
				return fmt.Errorf("failed to start gateway-service: %w", err)
			}

			fmt.Printf("  %s✓ Gateway started (port 8080)%s\n", colorGreen, colorReset)
			fmt.Printf("  %s壺（Tsubo）のエントリーポイント起動完了%s\n", colorGreen, colorReset)
			fmt.Println()

			// Add gateway to services list for log viewing
			services = append(services, "gateway-service")
		}
	}

	fmt.Printf("%s✓ All services started!%s\n", colorGreen, colorReset)
	fmt.Println()

	if *detachFlag {
		fmt.Println("Services are running in the background.")
		fmt.Println("To view logs: docker compose logs -f")
		fmt.Println("To stop services: docker compose down in each service directory")
	} else {
		// If not detached, show logs from all services
		fmt.Println("Showing logs from all services (Ctrl+C to stop)...")
		fmt.Println()

		// Collect container IDs from all services
		var containerIDs []string
		for _, service := range services {
			serviceDir := filepath.Join(implDir, service)
			composeFile := filepath.Join(serviceDir, "docker-compose.yml")
			if _, err := os.Stat(composeFile); err == nil {
				// Get container IDs using docker compose ps -q
				cmd := exec.Command("docker", "compose", "ps", "-q")
				cmd.Dir = serviceDir
				output, err := cmd.Output()
				if err == nil && len(output) > 0 {
					// Split by newline in case multiple containers per service
					ids := strings.Split(strings.TrimSpace(string(output)), "\n")
					for _, id := range ids {
						if id != "" {
							containerIDs = append(containerIDs, id)
						}
					}
				}
			}
		}

		// Show logs from all containers in parallel
		if len(containerIDs) > 0 {
			var wg sync.WaitGroup
			for _, containerID := range containerIDs {
				wg.Add(1)
				go func(id string) {
					defer wg.Done()
					cmd := exec.Command("docker", "logs", "-f", "--tail", "50", id)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Run()
				}(containerID)
			}
			wg.Wait()
		}
	}

	return nil
}

func printRunUsage() {
	fmt.Println("Usage: potter run [options] <tsubo-file>")
	fmt.Println()
	fmt.Println("Starts all services using docker-compose")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -d                Run in detached mode (background)")
	fmt.Println("  --service NAME    Run specific service only")
	fmt.Println("  --help            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter run ./poc/contracts/app.tsubo.yaml              # Start all services")
	fmt.Println("  potter run -d ./poc/contracts/app.tsubo.yaml           # Start in background")
	fmt.Println("  potter run --service user ./poc/contracts/app.tsubo.yaml  # Start user-service only")
}
