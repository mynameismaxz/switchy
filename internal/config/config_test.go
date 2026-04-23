package config

import (
	"os"
	"testing"
)

func withTempDir(t *testing.T, fn func()) {
	t.Helper()
	dir := t.TempDir()
	configDirOverride = dir
	t.Cleanup(func() { configDirOverride = "" })
	fn()
}

func TestProfilesRoundtrip(t *testing.T) {
	withTempDir(t, func() {
		pf, err := LoadProfiles()
		if err != nil {
			t.Fatal(err)
		}
		if len(pf.Profiles) != 0 {
			t.Error("expected empty profiles on first load")
		}

		pf.Profiles["local"] = map[string]string{"FOO": "bar"}
		if err := SaveProfiles(pf); err != nil {
			t.Fatal(err)
		}

		loaded, err := LoadProfiles()
		if err != nil {
			t.Fatal(err)
		}
		if loaded.Profiles["local"]["FOO"] != "bar" {
			t.Errorf("expected bar, got %q", loaded.Profiles["local"]["FOO"])
		}
		if loaded.Version != 1 {
			t.Errorf("expected version 1, got %d", loaded.Version)
		}
	})
}

func TestStateRoundtrip(t *testing.T) {
	withTempDir(t, func() {
		st, err := LoadState()
		if err != nil {
			t.Fatal(err)
		}
		if st.CurrentProfile != "" {
			t.Error("expected empty state on first load")
		}

		st.CurrentProfile = "prod"
		st.LastActivationMode = "persistent"
		if err := SaveState(st); err != nil {
			t.Fatal(err)
		}

		loaded, err := LoadState()
		if err != nil {
			t.Fatal(err)
		}
		if loaded.CurrentProfile != "prod" {
			t.Errorf("expected prod, got %q", loaded.CurrentProfile)
		}
	})
}

func TestFilePermissions(t *testing.T) {
	withTempDir(t, func() {
		pf := &ProfilesFile{Version: 1, Profiles: map[string]map[string]string{"x": {"K": "v"}}}
		if err := SaveProfiles(pf); err != nil {
			t.Fatal(err)
		}
		path, _ := ProfilesFilePath()
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("expected 0600 perms, got %v", info.Mode().Perm())
		}
	})
}
