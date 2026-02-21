package types

// ArchitectureDefinition represents an architecture definition file (.arch.yaml)
type ArchitectureDefinition struct {
	Version      string           `yaml:"version"`
	Architecture ArchitectureSpec `yaml:"architecture"`
}

// ArchitectureSpec contains the architecture specification details
type ArchitectureSpec struct {
	Name               string           `yaml:"name"`
	Description        string           `yaml:"description"`
	DirectoryStructure []DirectoryEntry `yaml:"directory_structure"`
	Rules              []string         `yaml:"rules"`
	Notes              string           `yaml:"notes"`
}

// DirectoryEntry represents a directory in the architecture's directory structure
type DirectoryEntry struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
}
