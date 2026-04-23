package profile

import (
	"fmt"
	"sort"

	"github.com/mynameismaxz/switchy/internal/config"
)

// UpsertProfile creates or updates a profile, merging kvPairs into existing keys.
func UpsertProfile(name string, kvPairs map[string]string) error {
	pf, err := config.LoadProfiles()
	if err != nil {
		return err
	}
	if pf.Profiles[name] == nil {
		pf.Profiles[name] = make(map[string]string)
	}
	for k, v := range kvPairs {
		pf.Profiles[name][k] = v
	}
	return config.SaveProfiles(pf)
}

// GetProfile returns the env var map for a named profile.
func GetProfile(name string) (map[string]string, error) {
	pf, err := config.LoadProfiles()
	if err != nil {
		return nil, err
	}
	vars, ok := pf.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return vars, nil
}

// DeleteProfile removes a profile. If force is false and the profile is current, returns an error.
func DeleteProfile(name string, force bool) error {
	pf, err := config.LoadProfiles()
	if err != nil {
		return err
	}
	if _, ok := pf.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	st, err := config.LoadState()
	if err != nil {
		return err
	}
	if st.CurrentProfile == name && !force {
		return fmt.Errorf("profile %q is the current profile\nUse --force to delete anyway", name)
	}
	delete(pf.Profiles, name)
	if err := config.SaveProfiles(pf); err != nil {
		return err
	}
	if st.CurrentProfile == name {
		st.CurrentProfile = ""
		st.LastActivationMode = ""
		return config.SaveState(st)
	}
	return nil
}

// ListProfiles returns sorted profile names and the current profile name.
func ListProfiles() (names []string, current string, err error) {
	pf, err := config.LoadProfiles()
	if err != nil {
		return nil, "", err
	}
	st, err := config.LoadState()
	if err != nil {
		return nil, "", err
	}
	for name := range pf.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, st.CurrentProfile, nil
}
