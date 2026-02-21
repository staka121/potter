package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/staka121/potter/internal/executor"
	"github.com/staka121/potter/pkg/state"
	"github.com/staka121/potter/pkg/types"
)

func runRefactor(args []string) error {
	fs := flag.NewFlagSet("refactor", flag.ExitOnError)
	serviceFlag := fs.String("service", "", "Specific service to refactor (default: all services)")
	concurrency := fs.Int("concurrency", 0, "Maximum parallel executions (0 = unlimited)")
	helpFlag := fs.Bool("help", false, "Show help for refactor command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printRefactorUsage()
		return nil
	}

	tsuboFile, err := parseTsuboFileArg(fs.Args())
	if err != nil {
		return err
	}

	tsubo, mgr, st, contractsDir, err := loadMigrateContext(tsuboFile)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s========================================%s\n", colorBlue, colorReset)
	fmt.Printf("%sPotter Refactor%s\n", colorBlue, colorReset)
	fmt.Printf("%s========================================%s\n\n", colorBlue, colorReset)

	// Determine which services to refactor
	var targets []types.ObjectRef
	if *serviceFlag != "" {
		found := false
		for _, obj := range tsubo.Objects {
			if obj.Name == *serviceFlag {
				targets = append(targets, obj)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("service %q not found in tsubo definition", *serviceFlag)
		}
		fmt.Printf("  Refactoring: %s%s%s\n\n", colorGreen, *serviceFlag, colorReset)
	} else {
		targets = tsubo.Objects
		fmt.Printf("  Refactoring all %d service(s)\n\n", len(targets))
	}

	implementationsDir := filepath.Join(filepath.Dir(tsuboFile), "implementations")
	projectRoot := filepath.Join(contractsDir, "..", "..")

	var changeRecords []types.ChangeRecord
	now := time.Now()

	for _, obj := range targets {
		fmt.Printf("  🔨 Refactoring: %s\n", obj.Name)

		plan := buildSingleServicePlanForRefactor(tsubo, tsuboFile, contractsDir, implementationsDir, projectRoot, obj)

		runner, err := executor.NewRunner(plan)
		if err != nil {
			return fmt.Errorf("failed to create runner for %s: %w", obj.Name, err)
		}
		if *concurrency > 0 {
			runner.SetConcurrency(*concurrency)
		}

		result, err := runner.ExecuteSingle(obj.Name)
		if err != nil {
			return fmt.Errorf("refactor failed for %s: %w", obj.Name, err)
		}
		if !result.Success {
			return fmt.Errorf("refactor failed for %s: %v", obj.Name, result.Error)
		}

		fmt.Printf("  ✅ %s refactored successfully\n\n", obj.Name)

		changeRecords = append(changeRecords, types.ChangeRecord{
			ServiceName: obj.Name,
			ChangeType:  "modified",
			Breaking:    false,
			Description: "Refactored from current contract",
		})

		// Update service state (hash stays the same; just bump migration version)
		if svcState, exists := st.Services[obj.Name]; exists {
			svcState.LastMigrated = now
			svcState.MigrationVersion++
		}
	}

	// Record refactor in migration history
	record := types.MigrationRecord{
		ID:          fmt.Sprintf("%d", now.UnixNano()),
		Timestamp:   now,
		Description: "Refactor: regenerated from current Contract",
		Type:        "refactor",
		Changes:     changeRecords,
	}
	st.Migrations = append([]types.MigrationRecord{record}, st.Migrations...)

	if err := mgr.Save(st); err != nil {
		fmt.Printf("%s⚠️  Warning: failed to update state: %v%s\n", colorYellow, err, colorReset)
	}

	fmt.Printf("%s✓ Refactor completed successfully!%s\n", colorGreen, colorReset)
	return nil
}

// buildSingleServicePlanForRefactor builds an ImplementationPlan for a single service refactor
func buildSingleServicePlanForRefactor(
	tsubo *types.TsuboDefinition,
	tsuboFile string,
	contractsDir string,
	implementationsDir string,
	projectRoot string,
	obj types.ObjectRef,
) *types.ImplementationPlan {
	wave := types.Wave{
		Wave:     0,
		Parallel: false,
		Objects: []types.ObjectInWave{
			{
				Name:         obj.Name,
				Contract:     filepath.Join(contractsDir, obj.Contract),
				Dependencies: obj.Dependencies,
				Port:         obj.Runtime.Port,
				IsGateway:    false,
			},
		},
	}

	return &types.ImplementationPlan{
		Tsubo:              tsubo.Tsubo.Name,
		TsuboFile:          tsuboFile,
		ContractsDir:       contractsDir,
		ProjectRoot:        projectRoot,
		ImplementationsDir: implementationsDir,
		ContextFiles:       getRefactorContextFiles(projectRoot),
		Waves:              []types.Wave{wave},
	}
}

// getRefactorContextFiles returns existing doc files
func getRefactorContextFiles(projectRoot string) []string {
	candidates := []string{
		filepath.Join(projectRoot, "docs", "PHILOSOPHY.md"),
		filepath.Join(projectRoot, "docs", "DEVELOPMENT_PRINCIPLES.md"),
		filepath.Join(projectRoot, "docs", "WHY_GO.md"),
		filepath.Join(projectRoot, "docs", "CONTRACT_DESIGN.md"),
	}

	var existing []string
	for _, f := range candidates {
		if _, err := os.Stat(f); err == nil {
			existing = append(existing, f)
		}
	}
	return existing
}

// loadStateForRefactor is a helper alias used in refactor.go (uses the shared loadMigrateContext)
var _ = func(tsuboFile string) (*state.Manager, *types.PotterState, error) {
	_, mgr, st, _, err := loadMigrateContext(tsuboFile)
	return mgr, st, err
}

func printRefactorUsage() {
	fmt.Println("Usage: potter refactor [options] <tsubo-file>")
	fmt.Println()
	fmt.Println("Refactor regenerates service implementations cleanly from the current Contract.")
	fmt.Println("Contract is the Source of Truth — refactor = remove patchwork, start fresh.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --service <name>   Refactor only this service (default: all services)")
	fmt.Println("  --concurrency N    Maximum parallel executions (default: unlimited)")
	fmt.Println("  --help             Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter refactor app.tsubo.yaml                      # Refactor all services")
	fmt.Println("  potter refactor --service todo-service app.tsubo.yaml  # Refactor one service")
}
