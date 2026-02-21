package types

import "time"

// PotterState represents the state of a potter-managed application
type PotterState struct {
	Version    string                   `json:"version"`
	TsuboFile  string                   `json:"tsubo_file"`
	TsuboHash  string                   `json:"tsubo_hash"`
	Services   map[string]*ServiceState `json:"services"`
	Migrations []MigrationRecord        `json:"migrations"`
}

// ServiceState represents the state of a single service
type ServiceState struct {
	ContractFile     string    `json:"contract_file"`
	ContractHash     string    `json:"contract_hash"`
	ContractSnapshot string    `json:"contract_snapshot"` // Full YAML for diff analysis
	LastMigrated     time.Time `json:"last_migrated"`
	MigrationVersion int       `json:"migration_version"`
}

// MigrationRecord represents a single migration event
type MigrationRecord struct {
	ID          string         `json:"id"`
	Timestamp   time.Time      `json:"timestamp"`
	Description string         `json:"description"`
	Type        string         `json:"type"` // "migrate" or "refactor"
	Changes     []ChangeRecord `json:"changes"`
}

// ChangeRecord represents a single change within a migration
type ChangeRecord struct {
	ServiceName string `json:"service_name"`
	ChangeType  string `json:"change_type"` // "added", "modified", "removed"
	Breaking    bool   `json:"breaking"`
	Description string `json:"description"`
}
