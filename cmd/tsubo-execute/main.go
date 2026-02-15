package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/staka121/tsubo/internal/executor"
	"github.com/staka121/tsubo/pkg/types"
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
		fmt.Println("Usage: tsubo-execute <plan.json>")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  tsubo-execute /tmp/tsubo-implementation-plan.json")
		return fmt.Errorf("invalid arguments")
	}

	planFile := os.Args[1]

	// Verify file exists
	if _, err := os.Stat(planFile); os.IsNotExist(err) {
		return fmt.Errorf("plan file not found: %s", planFile)
	}

	printHeader()

	// Load implementation plan
	fmt.Printf("%s[Step 1] Loading implementation plan%s\n", colorYellow, colorReset)
	plan, err := executor.LoadPlan(planFile)
	if err != nil {
		return err
	}

	fmt.Printf("  %sTsubo: %s%s\n", colorGreen, plan.Tsubo, colorReset)
	fmt.Printf("  Objects: %d\n", countObjects(plan))
	fmt.Printf("  Waves: %d\n", len(plan.Waves))
	fmt.Println()

	// Generate prompts
	fmt.Printf("%s[Step 2] Generating implementation prompts%s\n", colorYellow, colorReset)
	generator := executor.NewPromptGenerator(plan)

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
			promptFile := fmt.Sprintf("/tmp/tsubo-prompt-%s.md", obj.Name)
			if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
				return fmt.Errorf("failed to write prompt file: %w", err)
			}

			fmt.Printf("    Prompt saved: %s\n", promptFile)
		}
	}

	fmt.Println()

	// Summary
	printSummary(plan)

	return nil
}

func printHeader() {
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sTsubo Execution Engine%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Println()
}

func printSummary(plan *types.ImplementationPlan) {
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sPrompts Generated%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n", colorBlue, colorReset)
	fmt.Println()

	fmt.Printf("Tsubo: %s\n", plan.Tsubo)
	fmt.Printf("Total objects: %d\n", countObjects(plan))
	fmt.Println()

	fmt.Printf("%sNext steps:%s\n", colorGreen, colorReset)
	fmt.Println("1. Review the generated prompts in /tmp/tsubo-prompt-*.md")
	fmt.Println("2. Use AI agents (e.g., Claude Code Task tool) to implement each service")
	fmt.Println("3. Follow the wave order:")

	for _, wave := range plan.Waves {
		fmt.Printf("   Wave %d: %s\n", wave.Wave, getObjectNames(wave))
	}

	fmt.Println()
	fmt.Printf("%sImplementation guidelines:%s\n", colorYellow, colorReset)
	fmt.Println("- Each prompt contains full context (philosophy + contract)")
	fmt.Println("- AI agents should read the prompt and implement independently")
	fmt.Println("- Wave 0 objects can be implemented in parallel")
	fmt.Printf("- Wave 1+ objects must wait for previous wave completion\n")
	fmt.Println()
}

func countObjects(plan *types.ImplementationPlan) int {
	count := 0
	for _, wave := range plan.Waves {
		count += len(wave.Objects)
	}
	return count
}

func getObjectNames(wave types.Wave) string {
	var names []string
	for _, obj := range wave.Objects {
		names = append(names, obj.Name)
	}
	return strings.Join(names, ", ")
}
