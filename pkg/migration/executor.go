package migration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/staka121/potter/internal/executor"
	"github.com/staka121/potter/pkg/types"
)

// ExecuteMigration carries out all steps in a migration plan
func ExecuteMigration(
	plan *MigrationPlan,
	tsubo *types.TsuboDefinition,
	tsuboFile string,
	state *types.PotterState,
	apiKey string,
	concurrency int,
) error {
	contractsDir := filepath.Dir(tsuboFile)
	tsuboDir := filepath.Dir(tsuboFile)
	implementationsDir := filepath.Join(tsuboDir, "implementations")

	for _, step := range plan.Steps {
		switch step.Action {
		case "implement_new", "reimplement":
			fmt.Printf("\n  🔨 [%s] %s\n", step.Action, step.ServiceName)
			if err := executeServiceBuild(step.ServiceName, tsubo, tsuboFile, contractsDir, implementationsDir, concurrency); err != nil {
				return fmt.Errorf("failed to %s %s: %w", step.Action, step.ServiceName, err)
			}

		case "remove":
			fmt.Printf("\n  🗑️  [remove] %s\n", step.ServiceName)
			if err := removeServiceImpl(implementationsDir, step.ServiceName); err != nil {
				return fmt.Errorf("failed to remove %s: %w", step.ServiceName, err)
			}

		case "update_infra":
			fmt.Printf("\n  🏗️  [update_infra] Regenerating infrastructure\n")
			// Infrastructure is regenerated as part of the individual service builds.
			// The gateway service handles docker-compose orchestration.
			// This step is a no-op placeholder for future infra-only regeneration.
			fmt.Printf("      ✓ Infrastructure will be updated by service reimplementations\n")
		}
	}

	return nil
}

// executeServiceBuild builds a single service using the executor runner
func executeServiceBuild(
	serviceName string,
	tsubo *types.TsuboDefinition,
	tsuboFile string,
	contractsDir string,
	implementationsDir string,
	concurrency int,
) error {
	// Find the object definition
	var targetObj *types.ObjectRef
	for i, obj := range tsubo.Objects {
		if obj.Name == serviceName {
			targetObj = &tsubo.Objects[i]
			break
		}
	}

	if targetObj == nil {
		return fmt.Errorf("service %s not found in tsubo definition", serviceName)
	}

	// Build a minimal implementation plan for this single service
	plan := buildSingleServicePlan(tsubo, tsuboFile, contractsDir, implementationsDir, *targetObj)

	runner, err := executor.NewRunner(plan)
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	if concurrency > 0 {
		runner.SetConcurrency(concurrency)
	}

	result, err := runner.ExecuteSingle(serviceName)
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("implementation failed: %v", result.Error)
	}

	return nil
}

// buildSingleServicePlan creates a minimal ImplementationPlan for one service
func buildSingleServicePlan(
	tsubo *types.TsuboDefinition,
	tsuboFile string,
	contractsDir string,
	implementationsDir string,
	obj types.ObjectRef,
) *types.ImplementationPlan {
	projectRoot := filepath.Join(contractsDir, "..", "..")

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
		ContextFiles:       getExistingContextFiles(projectRoot),
		Waves:              []types.Wave{wave},
	}
}

// getExistingContextFiles returns context doc files that exist on disk
func getExistingContextFiles(projectRoot string) []string {
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

// removeServiceImpl deletes the implementation directory for a service
func removeServiceImpl(implementationsDir, serviceName string) error {
	serviceDir := filepath.Join(implementationsDir, serviceName)

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		fmt.Printf("      ⚠️  Implementation directory not found (already removed?): %s\n", serviceDir)
		return nil
	}

	if err := os.RemoveAll(serviceDir); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", serviceDir, err)
	}

	fmt.Printf("      ✓ Removed: %s\n", serviceDir)
	return nil
}
