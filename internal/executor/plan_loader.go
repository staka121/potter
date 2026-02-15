package executor

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/staka121/tsubo/pkg/types"
)

// LoadPlan loads an implementation plan from a JSON file
func LoadPlan(planFile string) (*types.ImplementationPlan, error) {
	data, err := os.ReadFile(planFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	var plan types.ImplementationPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	return &plan, nil
}
