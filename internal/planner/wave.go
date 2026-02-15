package planner

import (
	"github.com/staka121/tsubo/internal/analyzer"
	"github.com/staka121/tsubo/pkg/types"
)

// GenerateWaves creates implementation waves based on dependencies
func GenerateWaves(objects []analyzer.ObjectWithDeps) []types.Wave {
	// Simple two-wave strategy:
	// Wave 0: Objects with no dependencies
	// Wave 1: Objects with dependencies

	var wave0Objects []types.ObjectInWave
	var wave1Objects []types.ObjectInWave

	for _, obj := range objects {
		objInWave := types.ObjectInWave{
			Name:         obj.Name,
			Contract:     obj.Contract,
			Dependencies: obj.Dependencies,
		}

		if len(obj.Dependencies) == 0 {
			wave0Objects = append(wave0Objects, objInWave)
		} else {
			wave1Objects = append(wave1Objects, objInWave)
		}
	}

	var waves []types.Wave

	// Add Wave 0 if there are objects without dependencies
	if len(wave0Objects) > 0 {
		waves = append(waves, types.Wave{
			Wave:     0,
			Parallel: true,
			Objects:  wave0Objects,
		})
	}

	// Add Wave 1 if there are objects with dependencies
	if len(wave1Objects) > 0 {
		waves = append(waves, types.Wave{
			Wave:     1,
			Parallel: true,
			Objects:  wave1Objects,
		})
	}

	return waves
}
