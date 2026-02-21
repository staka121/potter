package diff

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/staka121/potter/internal/parser"
	"github.com/staka121/potter/pkg/state"
	"github.com/staka121/potter/pkg/types"
)

// ContractChange represents a detected change in a service contract
type ContractChange struct {
	ServiceName string
	ChangeType  string   // "added" | "modified_non_breaking" | "modified_breaking" | "removed"
	OldHash     string
	NewHash     string
	Details     []string // Human-readable descriptions
}

// DetectChanges compares current contracts against saved state and returns a list of changes
func DetectChanges(
	st *types.PotterState,
	tsubo *types.TsuboDefinition,
	contractsDir string,
	stateManager *state.Manager,
) ([]ContractChange, error) {
	var changes []ContractChange

	// Build a set of current service names
	currentServices := make(map[string]types.ObjectRef)
	for _, obj := range tsubo.Objects {
		currentServices[obj.Name] = obj
	}

	// Check for removed services (in state but not in tsubo)
	for name, svcState := range st.Services {
		if _, exists := currentServices[name]; !exists {
			changes = append(changes, ContractChange{
				ServiceName: name,
				ChangeType:  "removed",
				OldHash:     svcState.ContractHash,
				NewHash:     "",
				Details:     []string{fmt.Sprintf("Service %s has been removed from tsubo", name)},
			})
		}
	}

	// Check for added or modified services
	for name, obj := range currentServices {
		contractPath := resolveContractPath(contractsDir, obj.Contract)

		newHash, err := stateManager.ComputeHash(contractPath)
		if err != nil {
			return nil, fmt.Errorf("failed to hash contract for %s: %w", name, err)
		}

		existingState, exists := st.Services[name]
		if !exists {
			// New service
			changes = append(changes, ContractChange{
				ServiceName: name,
				ChangeType:  "added",
				OldHash:     "",
				NewHash:     newHash,
				Details:     []string{fmt.Sprintf("New service %s added to tsubo", name)},
			})
			continue
		}

		// Check if contract changed
		if existingState.ContractHash == newHash {
			continue // No change
		}

		// Contract changed — classify as breaking or non-breaking
		changeType, details, err := classifyChange(existingState.ContractSnapshot, contractPath)
		if err != nil {
			// If we can't parse, conservatively treat as breaking
			changeType = "modified_breaking"
			details = []string{fmt.Sprintf("Failed to analyze change: %v", err)}
		}

		changes = append(changes, ContractChange{
			ServiceName: name,
			ChangeType:  changeType,
			OldHash:     existingState.ContractHash,
			NewHash:     newHash,
			Details:     details,
		})
	}

	// For breaking changes, mark dependent services as affected too
	changes = markDependentServices(changes, tsubo)

	return changes, nil
}

// resolveContractPath resolves a contract file path relative to the contracts directory
func resolveContractPath(contractsDir, contractRef string) string {
	if filepath.IsAbs(contractRef) {
		return contractRef
	}
	return filepath.Join(contractsDir, contractRef)
}

// classifyChange determines whether a contract change is breaking or non-breaking
func classifyChange(oldSnapshot, newContractPath string) (changeType string, details []string, err error) {
	newData, err := os.ReadFile(newContractPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read new contract: %w", err)
	}

	oldObj, err := parser.ParseObjectYAML([]byte(oldSnapshot))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse old contract: %w", err)
	}

	newObj, err := parser.ParseObjectYAML(newData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse new contract: %w", err)
	}

	// Build endpoint ID maps
	oldEndpoints := make(map[string]types.Endpoint)
	for _, ep := range oldObj.API.Endpoints {
		oldEndpoints[ep.ID] = ep
	}
	newEndpoints := make(map[string]types.Endpoint)
	for _, ep := range newObj.API.Endpoints {
		newEndpoints[ep.ID] = ep
	}

	isBreaking := false

	// Check for removed endpoints (breaking)
	for id, oldEp := range oldEndpoints {
		if _, exists := newEndpoints[id]; !exists {
			isBreaking = true
			details = append(details, fmt.Sprintf("Endpoint removed: %s %s", oldEp.Method, oldEp.Path))
		}
	}

	// Check for modified endpoints
	for id, newEp := range newEndpoints {
		oldEp, existed := oldEndpoints[id]
		if !existed {
			// New endpoint added (non-breaking)
			details = append(details, fmt.Sprintf("Endpoint added: %s %s", newEp.Method, newEp.Path))
			continue
		}

		// Check for breaking changes in request schema (required field removed/changed)
		if hasBreakingRequestChange(oldEp.Request, newEp.Request) {
			isBreaking = true
			details = append(details, fmt.Sprintf("Breaking request change in endpoint: %s %s", newEp.Method, newEp.Path))
		}
	}

	// Check for type changes (simplified: any type removal is breaking)
	for typeName := range oldObj.Types {
		if _, exists := newObj.Types[typeName]; !exists {
			isBreaking = true
			details = append(details, fmt.Sprintf("Type removed: %s", typeName))
		}
	}

	if len(details) == 0 {
		details = append(details, "Contract updated (description or metadata changes)")
	}

	if isBreaking {
		return "modified_breaking", details, nil
	}
	return "modified_non_breaking", details, nil
}

// hasBreakingRequestChange checks if removing or changing required fields constitutes a breaking change
func hasBreakingRequestChange(oldReq, newReq map[string]interface{}) bool {
	if oldReq == nil {
		return false
	}
	oldRequired, _ := extractRequired(oldReq)
	newRequired, _ := extractRequired(newReq)

	// If any previously required field is now missing, it's breaking
	for _, field := range oldRequired {
		found := false
		for _, nf := range newRequired {
			if field == nf {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	return false
}

// extractRequired extracts field names from a request schema map
func extractRequired(req map[string]interface{}) (required []string, optional []string) {
	for key, val := range req {
		if key == "required" || key == "body" || key == "params" {
			continue
		}
		fieldMap, ok := val.(map[string]interface{})
		if !ok {
			continue
		}
		if req, exists := fieldMap["required"]; exists {
			if b, ok := req.(bool); ok && b {
				required = append(required, key)
				continue
			}
		}
		optional = append(optional, key)
	}
	return required, optional
}

// markDependentServices adds reimplement entries for services that depend on breaking-changed services
func markDependentServices(changes []ContractChange, tsubo *types.TsuboDefinition) []ContractChange {
	// Find all breaking changes
	breakingServices := make(map[string]bool)
	for _, ch := range changes {
		if ch.ChangeType == "modified_breaking" {
			breakingServices[ch.ServiceName] = true
		}
	}

	if len(breakingServices) == 0 {
		return changes
	}

	// Find services already covered
	alreadyCovered := make(map[string]bool)
	for _, ch := range changes {
		alreadyCovered[ch.ServiceName] = true
	}

	// Find dependents
	for _, obj := range tsubo.Objects {
		if alreadyCovered[obj.Name] {
			continue
		}
		for _, dep := range obj.Dependencies {
			if breakingServices[dep] {
				changes = append(changes, ContractChange{
					ServiceName: obj.Name,
					ChangeType:  "modified_breaking",
					OldHash:     "",
					NewHash:     "",
					Details:     []string{fmt.Sprintf("Affected by breaking change in dependency: %s", dep)},
				})
				alreadyCovered[obj.Name] = true
				break
			}
		}
	}

	return changes
}
