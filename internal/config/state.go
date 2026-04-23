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
