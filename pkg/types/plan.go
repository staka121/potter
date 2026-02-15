package types

// ImplementationPlan represents the complete implementation plan
type ImplementationPlan struct {
	Tsubo        string   `json:"tsubo"`
	TsuboFile    string   `json:"tsubo_file"`
	ContractsDir string   `json:"contracts_dir"`
	ProjectRoot  string   `json:"project_root"`
	ContextFiles []string `json:"context_files"`
	Waves        []Wave   `json:"waves"`
}

// Wave represents a group of objects that can be implemented in parallel
type Wave struct {
	Wave     int            `json:"wave"`
	Parallel bool           `json:"parallel"`
	Objects  []ObjectInWave `json:"objects"`
}

// ObjectInWave represents an object within a wave
type ObjectInWave struct {
	Name         string   `json:"name"`
	Contract     string   `json:"contract"`
	Dependencies []string `json:"dependencies"`
}
