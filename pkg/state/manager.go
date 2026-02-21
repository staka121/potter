package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/staka121/potter/pkg/types"
)

const stateVersion = "1"
const stateFileName = "state.json"
const stateDirName = ".potter"

// Manager manages potter state for a tsubo project
type Manager struct {
	tsuboFilePath string
}

// NewManager creates a new state manager for the given tsubo file
func NewManager(tsuboFilePath string) *Manager {
	return &Manager{tsuboFilePath: tsuboFilePath}
}

// GetStateDir returns the path to the .potter directory
func (m *Manager) GetStateDir() string {
	return filepath.Join(filepath.Dir(m.tsuboFilePath), stateDirName)
}

// getStatePath returns the full path to state.json
func (m *Manager) getStatePath() string {
	return filepath.Join(m.GetStateDir(), stateFileName)
}

// IsInitialized returns true if state.json exists
func (m *Manager) IsInitialized() bool {
	_, err := os.Stat(m.getStatePath())
	return err == nil
}

// Load reads and returns the current state
func (m *Manager) Load() (*types.PotterState, error) {
	data, err := os.ReadFile(m.getStatePath())
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state types.PotterState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// Save writes the state to disk
func (m *Manager) Save(state *types.PotterState) error {
	if err := os.MkdirAll(m.GetStateDir(), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	if err := os.WriteFile(m.getStatePath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// ComputeHash computes the SHA256 hash of a file's contents
func (m *Manager) ComputeHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file for hashing: %w", err)
	}

	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum), nil
}

// Initialize creates the initial state from the current contracts
func (m *Manager) Initialize(tsubo *types.TsuboDefinition, contractsDir string) (*types.PotterState, error) {
	tsuboHash, err := m.ComputeHash(m.tsuboFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to hash tsubo file: %w", err)
	}

	services := make(map[string]*types.ServiceState)
	for _, obj := range tsubo.Objects {
		contractPath := filepath.Join(contractsDir, obj.Contract)
		if !filepath.IsAbs(obj.Contract) {
			// Contract path may be relative to the tsubo file's directory
			contractPath = filepath.Join(contractsDir, obj.Contract)
		}

		hash, err := m.ComputeHash(contractPath)
		if err != nil {
			return nil, fmt.Errorf("failed to hash contract for %s: %w", obj.Name, err)
		}

		snapshot, err := os.ReadFile(contractPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read contract for %s: %w", obj.Name, err)
		}

		services[obj.Name] = &types.ServiceState{
			ContractFile:     obj.Contract,
			ContractHash:     hash,
			ContractSnapshot: string(snapshot),
			LastMigrated:     time.Now(),
			MigrationVersion: 0,
		}
	}

	state := &types.PotterState{
		Version:    stateVersion,
		TsuboFile:  m.tsuboFilePath,
		TsuboHash:  tsuboHash,
		Services:   services,
		Migrations: []types.MigrationRecord{},
	}

	if err := m.Save(state); err != nil {
		return nil, err
	}

	return state, nil
}
