package main

import (
	"fmt"
	"os"
)

const version = "0.5.0"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	command := os.Args[1]

	switch command {
	case "new":
		return runNew(os.Args[2:])
	case "build":
		return runBuild(os.Args[2:])
	case "verify":
		return runVerify(os.Args[2:])
	case "run":
		return runRun(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("potter version %s\n", version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printUsage() {
	fmt.Println("Potter - The Craftsman for Tsubo (AI-Driven Microservices)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  potter <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  new [service]              Create new service definition (default: example)")
	fmt.Println("  build <tsubo-file>         Generate implementation plan and execute")
	fmt.Println("  verify <tsubo-file>        Verify contract compliance and run tests")
	fmt.Println("  run [options] <tsubo-file> Start all services with docker-compose")
	fmt.Println("  version                Show version information")
	fmt.Println("  help                   Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter new                                   # Generate example service")
	fmt.Println("  potter new user-service                      # Create user-service template")
	fmt.Println("  potter build app.tsubo.yaml                  # AI-driven implementation (default)")
	fmt.Println("  potter build --concurrency 4 app.tsubo.yaml  # Limit parallel execution")
	fmt.Println("  potter build --prompt-only app.tsubo.yaml    # Generate prompts only")
	fmt.Println("  potter verify app.tsubo.yaml                 # Run contract verification")
	fmt.Println("  potter run -d app.tsubo.yaml                 # Start all services in background")
	fmt.Println()
}
