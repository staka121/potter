package migration

import (
	"github.com/staka121/potter/pkg/diff"
	"github.com/staka121/potter/pkg/types"
)

// MigrationPlan describes what actions need to be taken
type MigrationPlan struct {
	Changes     []diff.ContractChange
	Steps       []MigrationStep
	HasBreaking bool
}

// MigrationStep represents a single action in the migration plan
type MigrationStep struct {
	ServiceName string
	Action      string // "implement_new" | "reimplement" | "remove" | "update_infra"
	Description string
	Breaking    bool
}

// PlanMigration creates a migration plan from a set of detected changes
func PlanMigration(changes []diff.ContractChange, tsubo *types.TsuboDefinition) *MigrationPlan {
	plan := &MigrationPlan{
		Changes: changes,
	}

	// Track which services already have a step to avoid duplicates
	covered := make(map[string]bool)

	for _, ch := range changes {
		switch ch.ChangeType {
		case "added":
			plan.Steps = append(plan.Steps, MigrationStep{
				ServiceName: ch.ServiceName,
				Action:      "implement_new",
				Description: "New service — implement with AI",
				Breaking:    false,
			})
			covered[ch.ServiceName] = true

		case "modified_non_breaking", "modified_breaking":
			if !covered[ch.ServiceName] {
				isBreaking := ch.ChangeType == "modified_breaking"
				if isBreaking {
					plan.HasBreaking = true
				}
				plan.Steps = append(plan.Steps, MigrationStep{
					ServiceName: ch.ServiceName,
					Action:      "reimplement",
					Description: "Contract changed — re-implement with AI",
					Breaking:    isBreaking,
				})
				covered[ch.ServiceName] = true
			}

		case "removed":
			plan.Steps = append(plan.Steps, MigrationStep{
				ServiceName: ch.ServiceName,
				Action:      "remove",
				Description: "Service removed — delete implementation",
				Breaking:    true,
			})
			covered[ch.ServiceName] = true
			plan.HasBreaking = true
		}
	}

	// Always add infra update step at the end if there are any changes
	if len(changes) > 0 {
		plan.Steps = append(plan.Steps, MigrationStep{
			ServiceName: "",
			Action:      "update_infra",
			Description: "Regenerate infrastructure (docker-compose)",
			Breaking:    false,
		})
	}

	return plan
}
