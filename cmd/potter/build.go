package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/staka121/potter/internal/analyzer"
	"github.com/staka121/potter/internal/executor"
	"github.com/staka121/potter/internal/parser"
	"github.com/staka121/potter/internal/planner"
	"github.com/staka121/potter/pkg/types"
)

const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[0;34m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorRed    = "\033[0;31m"
)

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)

	promptOnlyFlag := fs.Bool("prompt-only", false, "Generate prompts only (skip AI implementation)")
	concurrency := fs.Int("concurrency", 0, "Maximum parallel executions (0 = unlimited)")
	helpFlag := fs.Bool("help", false, "Show help for build command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printBuildUsage()
		return nil
	}

	// Get tsubo file path
	args = fs.Args()
	if len(args) == 0 {
		return fmt.Errorf("tsubo file path required. Usage: potter build <tsubo-file> [options]")
	}

	tsuboFile := args[0]

	// Verify file exists
	if _, err := os.Stat(tsuboFile); os.IsNotExist(err) {
		return fmt.Errorf("tsubo file not found: %s", tsuboFile)
	}

	printBuildHeader()

	// Step 1: Parse tsubo file and generate plan
	fmt.Printf("%s[Step 1] Generating implementation plan%s\n", colorYellow, colorReset)
	plan, err := generatePlan(tsuboFile)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	fmt.Printf("  %sTsubo: %s%s\n", colorGreen, plan.Tsubo, colorReset)
	fmt.Printf("  Objects: %d\n", countObjects(plan))
	fmt.Printf("  Waves: %d\n", len(plan.Waves))
	fmt.Println()

	// Step 2: Generate prompts or execute with AI
	if *promptOnlyFlag {
		return generatePromptsOnly(plan)
	}

	return executeWithAI(plan, *concurrency)
}

func generatePlan(tsuboFile string) (*types.ImplementationPlan, error) {
	// Parse tsubo file
	tsuboDef, err := parser.ParseTsuboFile(tsuboFile)
	if err != nil {
		return nil, err
	}

	// Determine directories
	contractsDir := parser.GetContractsDir(tsuboFile)
	projectRoot := parser.GetProjectRoot(contractsDir)

	// Implementations directory is in the same directory as the tsubo file
	tsuboDir := filepath.Dir(tsuboFile)
	implementationsDir := filepath.Join(tsuboDir, "implementations")

	// Analyze dependencies
	objectsWithDeps, err := analyzer.AnalyzeDependencies(tsuboDef, contractsDir)
	if err != nil {
		return nil, err
	}

	// Generate waves
	waves := planner.GenerateWaves(objectsWithDeps)

	// Get context files
	contextFiles := planner.GetContextFiles(projectRoot)

	// Create implementation plan
	plan := &types.ImplementationPlan{
		Tsubo:              tsuboDef.Tsubo.Name,
		TsuboFile:          tsuboFile,
		ContractsDir:       contractsDir,
		ProjectRoot:        projectRoot,
		ImplementationsDir: implementationsDir,
		ContextFiles:       contextFiles,
		Waves:              waves,
	}

	return plan, nil
}

func executeWithAI(plan *types.ImplementationPlan, concurrency int) error {
	fmt.Printf("%s[Step 2] Executing with Claude API%s\n", colorYellow, colorReset)

	if concurrency > 0 {
		fmt.Printf("Concurrency limit: %d\n", concurrency)
	} else {
		fmt.Println("Concurrency: unlimited (wave-based parallelism)")
	}

	fmt.Printf("%sWARNING: This will use Claude API credits%s\n", colorYellow, colorReset)
	fmt.Println()

	// Create runner
	runner, err := executor.NewRunner(plan)
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	fmt.Printf("Temporary files will be saved to: %s\n", runner.GetTempDir())
	fmt.Println()

	// Set concurrency limit if specified
	if concurrency > 0 {
		runner.SetConcurrency(concurrency)
	}

	// Execute all waves
	results, err := runner.ExecuteAll()
	if err != nil {
		fmt.Printf("\n%sExecution failed: %v%s\n", colorRed, err, colorReset)
		executor.PrintSummary(results)
		return err
	}

	// Print summary
	executor.PrintSummary(results)

	fmt.Printf("\n%s✓ All implementations completed successfully!%s\n", colorGreen, colorReset)
	return nil
}

func generatePromptsOnly(plan *types.ImplementationPlan) error {
	fmt.Printf("%s[Step 2] Generating implementation prompts%s\n", colorYellow, colorReset)
	generator := executor.NewPromptGenerator(plan)

	// Create timestamped temp directory
	// Format: /tmp/potter/{app-name}/yyyymmddhhmmss
	timestamp := time.Now().Format("20060102150405")
	tempDir := filepath.Join("/tmp", "potter", plan.Tsubo, timestamp)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	for _, wave := range plan.Waves {
		fmt.Printf("\n%sWave %d:%s\n", colorBlue, wave.Wave, colorReset)
		fmt.Printf("Objects to implement: %d\n", len(wave.Objects))

		if wave.Parallel {
			fmt.Println("Execution mode: Parallel")
		} else {
			fmt.Println("Execution mode: Sequential")
		}
		fmt.Println()

		for _, obj := range wave.Objects {
			fmt.Printf("  %s- %s%s", colorGreen, obj.Name, colorReset)
			if len(obj.Dependencies) > 0 {
				fmt.Printf(" (depends on: %v)", obj.Dependencies)
			}
			fmt.Println()

			// Generate prompt
			prompt, err := generator.GeneratePrompt(obj)
			if err != nil {
				return fmt.Errorf("failed to generate prompt for %s: %w", obj.Name, err)
			}

			// Write prompt to file
			promptFile := filepath.Join(tempDir, fmt.Sprintf("tsubo-prompt-%s.md", obj.Name))
			if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
				return fmt.Errorf("failed to write prompt file: %w", err)
			}

			fmt.Printf("    Prompt saved: %s\n", promptFile)
		}
	}

	fmt.Println()
	fmt.Printf("%sPrompts saved to: %s%s\n", colorGreen, tempDir, colorReset)
	fmt.Println()
	printBuildSummary(plan)

	return nil
}

func printBuildHeader() {
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sPotter Build%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Println()
}

func printBuildUsage() {
	fmt.Println("Usage: potter build [options] <tsubo-file>")
	fmt.Println()
	fmt.Println("By default, potter build executes AI-driven implementation.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --prompt-only         Generate prompts only (skip AI implementation)")
	fmt.Println("  --concurrency N       Maximum parallel executions (default: unlimited)")
	fmt.Println("  --help                Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter build app.tsubo.yaml                    # AI-driven implementation")
	fmt.Println("  potter build --concurrency 4 app.tsubo.yaml    # Limit parallel execution")
	fmt.Println("  potter build --prompt-only app.tsubo.yaml      # Generate prompts only")
}

func printBuildSummary(plan *types.ImplementationPlan) {
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sPrompts Generated%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Println()

	fmt.Printf("Tsubo: %s\n", plan.Tsubo)
	fmt.Printf("Total objects: %d\n", countObjects(plan))
	fmt.Println()

	fmt.Printf("%sNext steps:%s\n", colorGreen, colorReset)
	fmt.Println("1. Review the generated prompts in /tmp/tsubo-prompt-*.md")
	fmt.Println("2. Execute with AI:")
	fmt.Println("   potter build <tsubo-file>")
	fmt.Println()
}

func countObjects(plan *types.ImplementationPlan) int {
	count := 0
	for _, wave := range plan.Waves {
		count += len(wave.Objects)
	}
	return count
}
