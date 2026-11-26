package activation

import (
	"strings"
	"testing"
)

func TestApplyScript_AlreadyInSync(t *testing.T) {
	cases := []struct {
		name string
		sync Sync
	}{
		{"zero generation", Sync{Generation: 0, AppliedGeneration: 0}},
		{"caught up", Sync{Generation: 5, AppliedGeneration: 5}},
		{"shell ahead (reset state)", Sync{Generation: 2, AppliedGeneration: 7}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ApplyScript(tc.sync); got != nil {
				t.Errorf("expected nil script, got %q", got)
			}
		})
	}
}

func TestApplyScript_FreshShellAppliesSelectedProfile(t *testing.T) {
	got := ApplyScript(Sync{
		Generation:     3,
		CurrentProfile: "local",
		CurrentVars:    map[string]string{"API_KEY": "dev-key", "BASE_URL": "http://localhost:8080"},
	})
	want := []string{
		"export API_KEY='dev-key'",
		"export BASE_URL='http://localhost:8080'",
		"export SWITCHY_PROFILE='local'",
		"export SWITCHY_APPLIED_GEN='3'",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestApplyScript_SwitchingProfilesUnsetsStaleKeys(t *testing.T) {
	got := ApplyScript(Sync{
		Generation:     4,
		CurrentProfile: "prod",
		CurrentVars:    map[string]string{"BASE_URL": "https://prod", "ONLY_IN_PROD": "x"},
		LoadedProfile:  "local",
		LoadedVars:     map[string]string{"BASE_URL": "http://localhost:8080", "ONLY_IN_LOCAL": "y"},
	})
	joined := strings.Join(got, "\n")
	wantIn := []string{
		"unset ONLY_IN_LOCAL",
		"export ONLY_IN_PROD='x'",
		"export BASE_URL='https://prod'",
		"export SWITCHY_PROFILE='prod'",
	}
	wantOut := []string{"unset BASE_URL", "export ONLY_IN_LOCAL"}
	for _, w := range wantIn {
		if !strings.Contains(joined, w) {
			t.Errorf("missing %q in:\n%s", w, joined)
		}
	}
	for _, w := range wantOut {
		if strings.Contains(joined, w) {
			t.Errorf("unexpected %q in:\n%s", w, joined)
		}
	}
	if !strings.HasSuffix(joined, "export SWITCHY_APPLIED_GEN='4'") {
		t.Errorf("generation marker must come last, got:\n%s", joined)
	}
}

func TestApplyScript_SameProfileNoStaleUnset(t *testing.T) {
	vars := map[string]string{"BASE_URL": "http://localhost:8080"}
	got := ApplyScript(Sync{
		Generation:     2,
		CurrentProfile: "local",
		CurrentVars:    vars,
		LoadedProfile:  "local",
		LoadedVars:     vars,
	})
	joined := strings.Join(got, "\n")
	if strings.Contains(joined, "unset") {
		t.Errorf("same profile should not unset anything, got:\n%s", joined)
	}
}

func TestApplyScript_UnloadedElsewhere(t *testing.T) {
	got := ApplyScript(Sync{
		Generation:    6,
		LoadedProfile: "local",
		LoadedVars:    map[string]string{"API_KEY": "dev-key"},
	})
	want := []string{
		"unset API_KEY SWITCHY_PROFILE",
		"export SWITCHY_APPLIED_GEN='6'",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestApplyScript_ClearedStateUnknownLoadedProfile(t *testing.T) {
	// Loaded profile was deleted from the store: only SWITCHY_PROFILE and the
	// generation marker can be written.
	got := ApplyScript(Sync{Generation: 7, LoadedProfile: "gone"})
	want := []string{
		"unset SWITCHY_PROFILE",
		"export SWITCHY_APPLIED_GEN='7'",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestApplyScript_ClearedStateNothingLoaded(t *testing.T) {
	got := ApplyScript(Sync{Generation: 8})
	want := []string{"export SWITCHY_APPLIED_GEN='8'"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestApplyScript_EscapesValues(t *testing.T) {
	got := ApplyScript(Sync{
		Generation:     1,
		CurrentProfile: "tricky",
		CurrentVars:    map[string]string{"MSG": "it's got a quote"},
	})
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, `export MSG='it'\''s got a quote'`) {
		t.Errorf("single quotes must be escaped, got:\n%s", joined)
	}
}

func TestOffScript(t *testing.T) {
	t.Run("nothing loaded", func(t *testing.T) {
		if got := OffScript(Off{}); got != nil {
			t.Errorf("expected nil, got %q", got)
		}
	})
	t.Run("loaded profile", func(t *testing.T) {
		got := OffScript(Off{
			LoadedProfile: "local",
			LoadedVars:    map[string]string{"API_KEY": "dev-key", "BASE_URL": "http://x"},
		})
		want := "unset API_KEY BASE_URL SWITCHY_PROFILE"
		if len(got) != 1 || got[0] != want {
			t.Errorf("got %q, want [%q]", got, want)
		}
	})
	t.Run("loaded profile missing from store", func(t *testing.T) {
		got := OffScript(Off{LoadedProfile: "gone"})
		want := "unset SWITCHY_PROFILE"
		if len(got) != 1 || got[0] != want {
			t.Errorf("got %q, want [%q]", got, want)
		}
	})
}
