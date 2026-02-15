package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	// Get tsubo file path (optional)
	remainingArgs := fs.Args()
	var implDir string

	if len(remainingArgs) > 0 {
		// Tsubo file specified
		tsuboFile := remainingArgs[0]
		tsuboDir := filepath.Dir(tsuboFile)
		implDir = filepath.Join(tsuboDir, "implementations")
	} else {
		// Default to poc/implementations for backward compatibility
		implDir = "poc/implementations"
	}

	if _, err := os.Stat(implDir); os.IsNotExist(err) {
		return fmt.Errorf("implementations directory not found: %s", implDir)
	}

	// Find all services
	entries, err := os.ReadDir(implDir)
	if err != nil {
		return fmt.Errorf("failed to read implementations directory: %w", err)
	}

	services := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			// Filter by service name if specified
			if *serviceFlag != "" && entry.Name() != *serviceFlag {
				continue
			}
			services = append(services, entry.Name())
		}
	}

	if len(services) == 0 {
		if *serviceFlag != "" {
			return fmt.Errorf("service not found: %s", *serviceFlag)
		}
		return fmt.Errorf("no services found in %s", implDir)
	}

	fmt.Printf("Starting %d service(s)...\n\n", len(services))

	// Start each service
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

		// Run docker-compose up
		cmdArgs := []string{"compose", "up"}
		if *detachFlag {
			cmdArgs = append(cmdArgs, "-d")
		}

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

	fmt.Printf("%s✓ All services started!%s\n", colorGreen, colorReset)
	fmt.Println()

	if *detachFlag {
		fmt.Println("Services are running in the background.")
		fmt.Println("To stop services, run: docker compose down in each service directory")
	}

	return nil
}

func printRunUsage() {
	fmt.Println("Usage: potter run [options] [tsubo-file]")
	fmt.Println()
	fmt.Println("Starts all services using docker-compose")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --service NAME    Run specific service only")
	fmt.Println("  -d                Run in detached mode (background)")
	fmt.Println("  --help            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter run                                      # Start all services (uses poc/implementations)")
	fmt.Println("  potter run -d                                   # Start all services (background)")
	fmt.Println("  potter run ./poc/contracts/app.tsubo.yaml       # Start services from tsubo file")
	fmt.Println("  potter run --service user                       # Start user-service only")
}
