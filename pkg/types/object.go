package types

// ObjectDefinition represents a single object (domain/microservice) contract
type ObjectDefinition struct {
	Version      string             `yaml:"version"`
	BelongsTo    string             `yaml:"belongs_to"`
	Service      ServiceConfig      `yaml:"service"`
	API          APIConfig          `yaml:"api"`
	Types        map[string]TypeDef `yaml:"types"`
	Dependencies DependenciesConfig `yaml:"dependencies"`
	Performance  PerformanceConfig  `yaml:"performance"`
}

// ServiceConfig contains service metadata
type ServiceConfig struct {
	Name         string         `yaml:"name"`
	Description  string         `yaml:"description"`
	Architecture string         `yaml:"architecture"`
	Context      ServiceContext `yaml:"context"`
}

// ServiceContext defines the business context
type ServiceContext struct {
	Purpose          string   `yaml:"purpose"`
	Domain           string   `yaml:"domain"`
	DomainBoundary   string   `yaml:"domain_boundary"`
	Responsibilities []string `yaml:"responsibilities"`
	Constraints      []string `yaml:"constraints"`
}

// APIConfig contains API definitions
type APIConfig struct {
	Version   string     `yaml:"version"`
	BasePath  string     `yaml:"base_path"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

// Endpoint represents an API endpoint
type Endpoint struct {
	ID       string                 `yaml:"id"`
	Method   string                 `yaml:"method"`
	Path     string                 `yaml:"path"`
	Request  map[string]interface{} `yaml:"request"`
	Response map[string]interface{} `yaml:"response"`
}

// TypeDef represents a type definition
type TypeDef struct {
	Description string                 `yaml:"description"`
	Properties  map[string]interface{} `yaml:"properties"`
}

// PerformanceConfig defines SLA latency requirements
type PerformanceConfig struct {
	Latency LatencyConfig `yaml:"latency"`
}

// LatencyConfig defines p50/p95/p99 latency thresholds (e.g. "50ms", "200ms")
type LatencyConfig struct {
	P50 string `yaml:"p50"`
	P95 string `yaml:"p95"`
	P99 string `yaml:"p99"`
}

// DependenciesConfig defines service dependencies
type DependenciesConfig struct {
	Services  []ServiceDependency  `yaml:"services"`
	Databases []DatabaseDependency `yaml:"databases"`
}

// ServiceDependency represents a dependency on another service
type ServiceDependency struct {
	Name      string   `yaml:"name"`
	Reason    string   `yaml:"reason"`
	Endpoints []string `yaml:"endpoints"`
	Type      string   `yaml:"type"`
}

// DatabaseDependency represents a database dependency
type DatabaseDependency struct {
	Name   string   `yaml:"name"`
	Type   string   `yaml:"type"`
	Tables []string `yaml:"tables"`
}
