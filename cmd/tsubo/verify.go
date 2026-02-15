package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	helpFlag := fs.Bool("help", false, "Show help for verify command")
	serviceFlag := fs.String("service", "", "Verify specific service only")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printVerifyUsage()
		return nil
	}

	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sTsubo Verify%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Println()

	implDir := "poc/implementations"
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

	fmt.Printf("Found %d service(s) to verify\n\n", len(services))

	// Verify each service
	passed := 0
	failed := 0

	for _, service := range services {
		fmt.Printf("%s[%s]%s\n", colorYellow, service, colorReset)

		serviceDir := filepath.Join(implDir, service)

		// Check for test script
		testScript := filepath.Join(serviceDir, "test.sh")
		if _, err := os.Stat(testScript); os.IsNotExist(err) {
			testScript = filepath.Join(serviceDir, "test-contract.sh")
		}

		if _, err := os.Stat(testScript); os.IsNotExist(err) {
			fmt.Printf("  %s⚠ No test script found (test.sh or test-contract.sh)%s\n", colorYellow, colorReset)
			fmt.Println()
			continue
		}

		// Run test script
		cmd := exec.Command("bash", testScript)
		cmd.Dir = serviceDir
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("  %s✗ Tests failed%s\n", colorRed, colorReset)
			fmt.Println(indent(string(output), "    "))
			fmt.Println()
			failed++
		} else {
			fmt.Printf("  %s✓ Tests passed%s\n", colorGreen, colorReset)
			fmt.Println()
			passed++
		}
	}

	// Summary
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sVerification Summary%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Println()
	fmt.Printf("Total services: %d\n", len(services))
	fmt.Printf("%sPassed: %d%s\n", colorGreen, passed, colorReset)

	if failed > 0 {
		fmt.Printf("%sFailed: %d%s\n", colorRed, failed, colorReset)
		return fmt.Errorf("verification failed for %d service(s)", failed)
	}

	fmt.Println()
	fmt.Printf("%s✓ All verifications passed!%s\n", colorGreen, colorReset)
	return nil
}

func indent(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, prefix+line)
		}
	}
	return strings.Join(result, "\n")
}

func printVerifyUsage() {
	fmt.Println("Usage: tsubo verify [options]")
	fmt.Println()
	fmt.Println("Verifies contract compliance and runs tests for all services")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --service NAME    Verify specific service only")
	fmt.Println("  --help            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tsubo verify                  # Verify all services")
	fmt.Println("  tsubo verify --service user   # Verify user-service only")
}
