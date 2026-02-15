package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/staka121/tsubo/internal/analyzer"
	"github.com/staka121/tsubo/internal/parser"
	"github.com/staka121/tsubo/internal/planner"
)

const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[0;34m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse arguments
	if len(os.Args) != 2 {
		fmt.Println("Usage: tsubo-plan <tsubo.yaml>")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  tsubo-plan ./poc/contracts/tsubo-todo-app.tsubo.yaml")
		return fmt.Errorf("invalid arguments")
	}

	tsuboFile := os.Args[1]

	// Verify file exists
	if _, err := os.Stat(tsuboFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", tsuboFile)
	}

	printHeader()

	// Step 0: Parse Tsubo file
	fmt.Printf("%s[Step 0] Parsing Tsubo file%s\n", colorYellow, colorReset)
	tsubo, err := parser.ParseTsuboFile(tsuboFile)
	if err != nil {
		return err
	}
	fmt.Printf("  %sTsubo: %s%s\n", colorGreen, tsubo.Tsubo.Name, colorReset)
	fmt.Println()

	// Get directories
	contractsDir := parser.GetContractsDir(tsuboFile)
	projectRoot := parser.GetProjectRoot(contractsDir)

	fmt.Printf("  Contracts directory: %s\n", contractsDir)
	fmt.Printf("  Project root: %s\n", projectRoot)
	fmt.Println()

	// Step 1: Verify context files
	fmt.Printf("%s[Step 1] Verifying context files%s\n", colorYellow, colorReset)
	verifyContextFiles(projectRoot)
	fmt.Println()

	// Step 2: Enumerate objects
	fmt.Printf("%s[Step 2] Enumerating objects%s\n", colorYellow, colorReset)
	fmt.Printf("Found %d object(s):\n", len(tsubo.Objects))
	for _, obj := range tsubo.Objects {
		fmt.Printf("  - %s\n", obj.Contract)
	}
	fmt.Println()

	// Step 3: Analyze dependencies
	fmt.Printf("%s[Step 3] Analyzing dependencies%s\n", colorYellow, colorReset)
	objects, err := analyzer.AnalyzeDependencies(tsubo, contractsDir)
	if err != nil {
		return err
	}

	for _, obj := range objects {
		if len(obj.Dependencies) == 0 {
			fmt.Printf("  %s: no dependencies\n", obj.Name)
		} else {
			fmt.Printf("  %s: depends on %v\n", obj.Name, obj.Dependencies)
		}
	}
	fmt.Println()

	// Step 4: Determine implementation order
	fmt.Printf("%s[Step 4] Determining implementation order%s\n", colorYellow, colorReset)
	waves := planner.GenerateWaves(objects)

	for _, wave := range waves {
		if wave.Wave == 0 {
			fmt.Printf("Wave 0 (parallel execution - no dependencies):\n")
		} else {
			fmt.Printf("Wave %d (execute after Wave %d completes):\n", wave.Wave, wave.Wave-1)
		}

		for _, obj := range wave.Objects {
			if len(obj.Dependencies) == 0 {
				fmt.Printf("  - %s\n", obj.Name)
			} else {
				fmt.Printf("  - %s (depends on: %v)\n", obj.Name, obj.Dependencies)
			}
		}
	}
	fmt.Println()

	// Step 5: Generate implementation plan
	fmt.Printf("%s[Step 5] Generating implementation plan%s\n", colorYellow, colorReset)
	plan := planner.GeneratePlan(tsubo, tsuboFile, contractsDir, projectRoot, objects)

	// Write plan to file
	planFile := "/tmp/tsubo-implementation-plan.json"
	planData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	if err := os.WriteFile(planFile, planData, 0644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	fmt.Printf("  %sImplementation plan generated: %s%s\n", colorGreen, planFile, colorReset)
	fmt.Println()

	// Summary
	printSummary(tsubo.Tsubo.Name, len(tsubo.Objects), planFile)

	return nil
}

func printHeader() {
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sTsubo Implementation Orchestrator%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Println()
}

func printSummary(tsuboName string, objectCount int, planFile string) {
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sReady for Implementation%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Println()
	fmt.Printf("Tsubo: %s\n", tsuboName)
	fmt.Printf("Number of objects: %d\n", objectCount)
	fmt.Printf("Implementation plan: %s\n", planFile)
	fmt.Println()
	fmt.Printf("%sNext steps:%s\n", colorGreen, colorReset)
	fmt.Println("1. Review the plan: cat", planFile, "| jq")
	fmt.Println("2. Start parallel implementation with AI agents")
	fmt.Println()
	fmt.Printf("%sEach AI agent will receive:%s\n", colorYellow, colorReset)
	fmt.Println("  - Tsubo philosophy (PHILOSOPHY.md)")
	fmt.Println("  - Development principles (DEVELOPMENT_PRINCIPLES.md)")
	fmt.Println("  - Why Go language (WHY_GO.md)")
	fmt.Println("  - Contract design (CONTRACT_DESIGN.md)")
	fmt.Println("  - Object contract (.object.yaml)")
	fmt.Println()
}

func verifyContextFiles(projectRoot string) {
	contextFiles := []string{
		"docs/PHILOSOPHY.md",
		"docs/DEVELOPMENT_PRINCIPLES.md",
		"docs/WHY_GO.md",
		"docs/CONTRACT_DESIGN.md",
	}

	for _, file := range contextFiles {
		fullPath := filepath.Join(projectRoot, file)
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("  ✓ %s\n", filepath.Base(file))
		} else {
			fmt.Printf("  ✗ %s not found\n", filepath.Base(file))
		}
	}
}
