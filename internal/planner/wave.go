package planner

import (
	"fmt"

	"github.com/staka121/tsubo/internal/analyzer"
	"github.com/staka121/tsubo/pkg/types"
)

// GenerateWaves creates implementation waves based on dependencies using topological sort
func GenerateWaves(objects []analyzer.ObjectWithDeps) []types.Wave {
	// Calculate depth for each object based on dependency graph
	depths := calculateDepths(objects)

	// Group objects by depth into waves
	waveMap := make(map[int][]types.ObjectInWave)
	maxDepth := 0

	for _, obj := range objects {
		depth := depths[obj.Name]
		if depth > maxDepth {
			maxDepth = depth
		}

		objInWave := types.ObjectInWave{
			Name:         obj.Name,
			Contract:     obj.Contract,
			Dependencies: obj.Dependencies,
		}

		waveMap[depth] = append(waveMap[depth], objInWave)
	}

	// Create waves in order
	var waves []types.Wave
	for i := 0; i <= maxDepth; i++ {
		if objects, exists := waveMap[i]; exists {
			waves = append(waves, types.Wave{
				Wave:     i,
				Parallel: true, // Objects in the same wave can run in parallel
				Objects:  objects,
			})
		}
	}

	return waves
}

// calculateDepths computes the depth of each object in the dependency graph
// Depth = longest path from any root node (object with no dependencies)
func calculateDepths(objects []analyzer.ObjectWithDeps) map[string]int {
	depths := make(map[string]int)
	visited := make(map[string]bool)
	inProgress := make(map[string]bool)

	// Build object map for quick lookup
	objMap := make(map[string]analyzer.ObjectWithDeps)
	for _, obj := range objects {
		objMap[obj.Name] = obj
	}

	// Calculate depth for each object using DFS
	var calculateDepth func(string) (int, error)
	calculateDepth = func(name string) (int, error) {
		// Check for cycles
		if inProgress[name] {
			return 0, fmt.Errorf("circular dependency detected: %s", name)
		}

		// Return cached result
		if visited[name] {
			return depths[name], nil
		}

		inProgress[name] = true
		defer func() { inProgress[name] = false }()

		obj, exists := objMap[name]
		if !exists {
			// Dependency not found in the object list (external dependency)
			// Treat as depth 0 (already implemented)
			depths[name] = 0
			visited[name] = true
			return 0, nil
		}

		// Base case: no dependencies
		if len(obj.Dependencies) == 0 {
			depths[name] = 0
			visited[name] = true
			return 0, nil
		}

		// Recursive case: depth = max(dependency depths) + 1
		maxDepth := -1
		for _, dep := range obj.Dependencies {
			depDepth, err := calculateDepth(dep)
			if err != nil {
				return 0, err
			}
			if depDepth > maxDepth {
				maxDepth = depDepth
			}
		}

		depth := maxDepth + 1
		depths[name] = depth
		visited[name] = true
		return depth, nil
	}

	// Calculate depth for all objects
	for _, obj := range objects {
		if !visited[obj.Name] {
			_, err := calculateDepth(obj.Name)
			if err != nil {
				// In case of cycle, assign to wave 0 (fallback)
				// This shouldn't happen with proper contract validation
				fmt.Printf("Warning: %v\n", err)
				depths[obj.Name] = 0
			}
		}
	}

	return depths
}
