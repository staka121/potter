package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/staka121/potter/internal/parser"
	"github.com/staka121/potter/pkg/diff"
	"github.com/staka121/potter/pkg/migration"
	"github.com/staka121/potter/pkg/state"
	"github.com/staka121/potter/pkg/types"
)

func runMigrate(args []string) error {
	if len(args) == 0 {
		printMigrateUsage()
		return nil
	}

	subcommand := args[0]
	rest := args[1:]

	switch subcommand {
	case "plan":
		return runMigratePlan(rest)
	case "apply":
		return runMigrateApply(rest)
	case "history":
		return runMigrateHistory(rest)
	case "help", "--help", "-h":
		printMigrateUsage()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown migrate subcommand: %s\n\n", subcommand)
		printMigrateUsage()
		return fmt.Errorf("unknown migrate subcommand: %s", subcommand)
	}
}

// runMigratePlan detects changes and displays the migration plan (dry run)
func runMigratePlan(args []string) error {
	tsuboFile, err := parseTsuboFileArg(args)
	if err != nil {
		return err
	}

	tsubo, mgr, st, contractsDir, err := loadMigrateContext(tsuboFile)
	if err != nil {
		return err
	}

	changes, err := diff.DetectChanges(st, tsubo, contractsDir, mgr)
	if err != nil {
		return fmt.Errorf("failed to detect changes: %w", err)
	}

	if len(changes) == 0 {
		fmt.Printf("\n✓ No changes detected in %s\n", filepath.Base(tsuboFile))
		return nil
	}

	plan := migration.PlanMigration(changes, tsubo)
	printMigrationPlan(plan, tsubo)

	fmt.Println("\nRun `potter migrate apply <tsubo-file>` to proceed.")
	return nil
}

// runMigrateApply executes the migration plan
func runMigrateApply(args []string) error {
	fs := flag.NewFlagSet("migrate apply", flag.ExitOnError)
	concurrency := fs.Int("concurrency", 0, "Maximum parallel executions (0 = unlimited)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	tsuboFile, err := parseTsuboFileArg(fs.Args())
	if err != nil {
		return err
	}

	tsubo, mgr, st, contractsDir, err := loadMigrateContext(tsuboFile)
	if err != nil {
		return err
	}

	changes, err := diff.DetectChanges(st, tsubo, contractsDir, mgr)
	if err != nil {
		return fmt.Errorf("failed to detect changes: %w", err)
	}

	if len(changes) == 0 {
		fmt.Printf("\n✓ No changes detected in %s\n", filepath.Base(tsuboFile))
		return nil
	}

	plan := migration.PlanMigration(changes, tsubo)
	printMigrationPlan(plan, tsubo)

	// Warn and confirm if breaking changes exist
	if plan.HasBreaking {
		fmt.Printf("\n%s⚠️  WARNING: Breaking changes detected!%s\n", colorRed, colorReset)
		fmt.Println("  Dependent services will be re-implemented.")
		fmt.Print("\nProceed? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Printf("\n%s[Migrate Apply] Executing migration plan...%s\n", colorYellow, colorReset)

	if err := migration.ExecuteMigration(plan, tsubo, tsuboFile, st, "", *concurrency); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Update state after successful migration
	if err := updateStateAfterMigration(mgr, st, tsubo, contractsDir, changes, "migrate"); err != nil {
		fmt.Printf("%s⚠️  Warning: failed to update state: %v%s\n", colorYellow, err, colorReset)
	}

	fmt.Printf("\n%s✓ Migration completed successfully!%s\n", colorGreen, colorReset)
	return nil
}

// runMigrateHistory displays migration history
func runMigrateHistory(args []string) error {
	tsuboFile, err := parseTsuboFileArg(args)
	if err != nil {
		return err
	}

	mgr := state.NewManager(tsuboFile)
	if !mgr.IsInitialized() {
		fmt.Println("No migration history (state not initialized).")
		fmt.Println("Run `potter migrate plan <tsubo-file>` to initialize.")
		return nil
	}

	st, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if len(st.Migrations) == 0 {
		fmt.Println("No migration history recorded yet.")
		return nil
	}

	fmt.Printf("\n%sMigration History for %s%s\n", colorBlue, filepath.Base(tsuboFile), colorReset)
	fmt.Println(strings.Repeat("─", 60))

	for i, rec := range st.Migrations {
		idx := len(st.Migrations) - i
		typeLabel := "[migrate]"
		if rec.Type == "refactor" {
			typeLabel = "[refactor]"
		}
		fmt.Printf("\n  #%d  %s  %s\n", idx, typeLabel, rec.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("      ID: %s\n", rec.ID)
		if rec.Description != "" {
			fmt.Printf("      %s\n", rec.Description)
		}
		for _, ch := range rec.Changes {
			breakingMark := ""
			if ch.Breaking {
				breakingMark = " (breaking)"
			}
			fmt.Printf("      • %s: %s%s\n", ch.ServiceName, ch.ChangeType, breakingMark)
		}
	}
	fmt.Println()
	return nil
}

// loadMigrateContext loads the tsubo definition, state manager, and state.
// If state is not initialized, it initializes it from current contracts.
func loadMigrateContext(tsuboFile string) (
	tsubo *types.TsuboDefinition,
	mgr *state.Manager,
	st *types.PotterState,
	contractsDir string,
	err error,
) {
	if _, err = os.Stat(tsuboFile); os.IsNotExist(err) {
		return nil, nil, nil, "", fmt.Errorf("tsubo file not found: %s", tsuboFile)
	}

	tsubo, err = parser.ParseTsuboFile(tsuboFile)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("failed to parse tsubo file: %w", err)
	}

	contractsDir = filepath.Dir(tsuboFile)
	mgr = state.NewManager(tsuboFile)

	if !mgr.IsInitialized() {
		fmt.Printf("%s[Info] State not found — initializing from current contracts...%s\n", colorYellow, colorReset)
		st, err = mgr.Initialize(tsubo, contractsDir)
		if err != nil {
			return nil, nil, nil, "", fmt.Errorf("failed to initialize state: %w", err)
		}
		fmt.Printf("  ✓ State initialized in %s\n", mgr.GetStateDir())
		fmt.Printf("  ✓ Tracking %d service(s)\n\n", len(st.Services))
	} else {
		st, err = mgr.Load()
		if err != nil {
			return nil, nil, nil, "", fmt.Errorf("failed to load state: %w", err)
		}
	}

	return tsubo, mgr, st, contractsDir, nil
}

// printMigrationPlan prints the migration plan in human-readable format
func printMigrationPlan(plan *migration.MigrationPlan, tsubo *types.TsuboDefinition) {
	fmt.Printf("\n%sDetected changes in %s:%s\n", colorBlue, tsubo.Tsubo.Name, colorReset)
	fmt.Println()

	for _, ch := range plan.Changes {
		switch ch.ChangeType {
		case "added":
			fmt.Printf("  %s[+] %s (NEW)%s\n", colorGreen, ch.ServiceName, colorReset)
		case "modified_non_breaking":
			fmt.Printf("  %s[~] %s (MODIFIED)%s\n", colorYellow, ch.ServiceName, colorReset)
		case "modified_breaking":
			fmt.Printf("  %s[~] %s (MODIFIED - BREAKING)%s\n", colorRed, ch.ServiceName, colorReset)
		case "removed":
			fmt.Printf("  %s[-] %s (REMOVED)%s\n", colorRed, ch.ServiceName, colorReset)
		}
		for _, detail := range ch.Details {
			fmt.Printf("      → %s\n", detail)
		}
	}

	fmt.Println()
	fmt.Printf("  %sMigration steps:%s\n", colorBlue, colorReset)
	for _, step := range plan.Steps {
		switch step.Action {
		case "implement_new":
			fmt.Printf("  [*] Implement new service: %s\n", step.ServiceName)
		case "reimplement":
			breakMark := ""
			if step.Breaking {
				breakMark = " (breaking)"
			}
			fmt.Printf("  [*] Re-implement: %s%s\n", step.ServiceName, breakMark)
		case "remove":
			fmt.Printf("  [*] Remove implementation: %s\n", step.ServiceName)
		case "update_infra":
			fmt.Printf("  [*] Update infrastructure (docker-compose)\n")
		}
	}
}

// updateStateAfterMigration updates state.json after a successful migration
func updateStateAfterMigration(
	mgr *state.Manager,
	st *types.PotterState,
	tsubo *types.TsuboDefinition,
	contractsDir string,
	changes []diff.ContractChange,
	migrationType string,
) error {
	now := time.Now()

	// Build migration record
	record := types.MigrationRecord{
		ID:        fmt.Sprintf("%d", now.UnixNano()),
		Timestamp: now,
		Type:      migrationType,
	}

	// Update service states and record changes
	for _, ch := range changes {
		changeRecord := types.ChangeRecord{
			ServiceName: ch.ServiceName,
			ChangeType:  ch.ChangeType,
			Breaking:    ch.ChangeType == "modified_breaking" || ch.ChangeType == "removed",
		}
		if len(ch.Details) > 0 {
			changeRecord.Description = ch.Details[0]
		}
		record.Changes = append(record.Changes, changeRecord)

		// Update or remove service state
		switch ch.ChangeType {
		case "removed":
			delete(st.Services, ch.ServiceName)

		case "added", "modified_non_breaking", "modified_breaking":
			// Find the object ref
			var objRef *types.ObjectRef
			for i, obj := range tsubo.Objects {
				if obj.Name == ch.ServiceName {
					objRef = &tsubo.Objects[i]
					break
				}
			}
			if objRef == nil {
				continue
			}

			contractPath := filepath.Join(contractsDir, objRef.Contract)
			newHash, err := mgr.ComputeHash(contractPath)
			if err != nil {
				continue
			}
			snapshot, err := os.ReadFile(contractPath)
			if err != nil {
				continue
			}

			existing := st.Services[ch.ServiceName]
			newVersion := 1
			if existing != nil {
				newVersion = existing.MigrationVersion + 1
			}

			st.Services[ch.ServiceName] = &types.ServiceState{
				ContractFile:     objRef.Contract,
				ContractHash:     newHash,
				ContractSnapshot: string(snapshot),
				LastMigrated:     now,
				MigrationVersion: newVersion,
			}
		}
	}

	// Update tsubo hash
	tsuboHash, _ := mgr.ComputeHash(mgr.GetStateDir() + "/../" + filepath.Base(st.TsuboFile))
	if tsuboHash != "" {
		st.TsuboHash = tsuboHash
	}

	// Prepend the new record (most recent first)
	st.Migrations = append([]types.MigrationRecord{record}, st.Migrations...)

	return mgr.Save(st)
}

// parseTsuboFileArg extracts the tsubo file path from args
func parseTsuboFileArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("tsubo file path required")
	}
	abs, err := filepath.Abs(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	return abs, nil
}

// stateToJSON is a helper for debug output
func stateToJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func printMigrateUsage() {
	fmt.Println("Usage: potter migrate <subcommand> <tsubo-file> [options]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  plan    <tsubo-file>   Detect changes and show migration plan (dry run)")
	fmt.Println("  apply   <tsubo-file>   Apply the migration plan")
	fmt.Println("  history <tsubo-file>   Show migration history")
	fmt.Println()
	fmt.Println("Options (apply):")
	fmt.Println("  --concurrency N        Maximum parallel executions (default: unlimited)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter migrate plan    poc/contracts/app.tsubo.yaml")
	fmt.Println("  potter migrate apply   poc/contracts/app.tsubo.yaml")
	fmt.Println("  potter migrate history poc/contracts/app.tsubo.yaml")
}

// Suppress unused warning — stateToJSON is kept for potential debug use
var _ = stateToJSON
