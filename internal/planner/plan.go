package planner

import (
	"os"
	"path/filepath"

	"github.com/staka121/tsubo/internal/analyzer"
	"github.com/staka121/tsubo/pkg/types"
)

// GeneratePlan creates a complete implementation plan
func GeneratePlan(
	tsubo *types.TsuboDefinition,
	tsuboFile string,
	contractsDir string,
	projectRoot string,
	objects []analyzer.ObjectWithDeps,
) *types.ImplementationPlan {
	// Get context files
	contextFiles := getContextFiles(projectRoot)

	// Generate waves
	waves := GenerateWaves(objects)

	return &types.ImplementationPlan{
		Tsubo:        tsubo.Tsubo.Name,
		TsuboFile:    tsuboFile,
		ContractsDir: contractsDir,
		ProjectRoot:  projectRoot,
		ContextFiles: contextFiles,
		Waves:        waves,
	}
}

// getContextFiles returns the list of context files that AI agents should read
func getContextFiles(projectRoot string) []string {
	contextFiles := []string{
		filepath.Join(projectRoot, "PHILOSOPHY.md"),
		filepath.Join(projectRoot, "docs", "DEVELOPMENT_PRINCIPLES.md"),
		filepath.Join(projectRoot, "docs", "WHY_GO.md"),
		filepath.Join(projectRoot, "docs", "CONTRACT_DESIGN.md"),
	}

	// Filter out files that don't exist
	var existing []string
	for _, file := range contextFiles {
		if _, err := os.Stat(file); err == nil {
			existing = append(existing, file)
		}
	}

	return existing
}
