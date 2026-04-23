package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type ProfilesFile struct {
	Version  int                            `json:"version"`
	Profiles map[string]map[string]string   `json:"profiles"`
}

func LoadProfiles() (*ProfilesFile, error) {
	path, err := ProfilesFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path) // #nosec G304 -- path is derived from os.UserHomeDir() + hardcoded filename, never from user input
	if errors.Is(err, os.ErrNotExist) {
		return &ProfilesFile{Version: 1, Profiles: make(map[string]map[string]string)}, nil
	}
	if err != nil {
		return nil, err
	}
	var pf ProfilesFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("malformed profiles.json: %w", err)
	}
	if pf.Profiles == nil {
		pf.Profiles = make(map[string]map[string]string)
	}
	return &pf, nil
}

func SaveProfiles(pf *ProfilesFile) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	path, err := ProfilesFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0600)
}

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
