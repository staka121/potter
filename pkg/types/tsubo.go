package types

// TsuboDefinition represents the entire application (tsubo)
type TsuboDefinition struct {
	Version string       `yaml:"version"`
	Tsubo   TsuboConfig  `yaml:"tsubo"`
	Objects []ObjectRef  `yaml:"objects"`
}

// TsuboConfig contains the tsubo metadata
type TsuboConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Purpose     string `yaml:"purpose"`
}

// ObjectRef references an object (domain/microservice) in the tsubo
type ObjectRef struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Contract    string   `yaml:"contract"`
	Runtime     Runtime  `yaml:"runtime"`
	Dependencies []string `yaml:"dependencies"`
}

// Runtime defines how the object runs
type Runtime struct {
	Type        string `yaml:"type"`
	Port        int    `yaml:"port"`
	HealthCheck string `yaml:"health_check"`
}
