package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type StateFile struct {
	CurrentProfile     string `json:"currentProfile"`
	LastActivationMode string `json:"lastActivationMode"` // "session" | "persistent" | ""
	// Generation is bumped when a profile is selected or turned off. The shell hook
	// compares it with $SWITCHY_APPLIED_GEN to decide whether a shell needs
	// to re-apply the selected profile. Old state files without this field
	// load as 0.
	Generation int `json:"generation,omitempty"`
}

// BumpGeneration increments Generation and saves the state file.
func (sf *StateFile) BumpGeneration() error {
	sf.Generation++
	return SaveState(sf)
}

func LoadState() (*StateFile, error) {
	path, err := StateFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path) // #nosec G304 -- path is derived from os.UserHomeDir() + hardcoded filename, never from user input
	if errors.Is(err, os.ErrNotExist) {
		return &StateFile{}, nil
	}
	if err != nil {
		return nil, err
	}
	var sf StateFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("malformed state.json: %w", err)
	}
	return &sf, nil
}

func SaveState(sf *StateFile) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	path, err := StateFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0600)
}
