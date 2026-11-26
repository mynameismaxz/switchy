// Package activation renders shell scripts that bring a shell's environment in
// sync with the profile selected in switchy state. It powers the shell hook
// installed by `swy init`: the hook evals the output of `swy hook-apply` on
// every prompt, so `swy use` / `swy off` take effect in existing shells
// without pasting or re-sourcing anything.
//
// The functions here are pure: cmd/ loads state, profiles and environment
// variables, fills in a Sync (or Off) value, and evals the returned lines.
package activation

import (
	"sort"
	"strconv"
	"strings"

	"github.com/mynameismaxz/switchy/internal/shell"
)

// AppliedGenEnv is the environment variable each hooked shell uses to remember
// which state generation it has last applied.
const AppliedGenEnv = "SWITCHY_APPLIED_GEN"

// ProfileEnv is the environment variable that marks which profile a shell has
// currently applied (also exported to user programs by `swy export`).
const ProfileEnv = "SWITCHY_PROFILE"

// Sync describes the gap between one shell session and the switchy state.
type Sync struct {
	// Generation is the current generation recorded in state.json.
	Generation int
	// AppliedGeneration is the last generation this shell applied
	// (parsed from $SWITCHY_APPLIED_GEN; 0 when unset).
	AppliedGeneration int
	// CurrentProfile is the profile selected in switchy state ("" when none).
	CurrentProfile string
	// CurrentVars holds the variables of CurrentProfile; nil when the profile
	// no longer exists.
	CurrentVars map[string]string
	// LoadedProfile is the profile this shell last applied
	// (parsed from $SWITCHY_PROFILE; "" when unset).
	LoadedProfile string
	// LoadedVars holds the variables of LoadedProfile; nil when the profile
	// no longer exists.
	LoadedVars map[string]string
}

// ApplyScript returns the shell lines this session must eval to catch up with
// the current state, or nil when the shell is already in sync.
//
// Cases, in order:
//   - state generation reached → nothing to do
//   - a profile is selected    → export its variables (dropping stale keys
//     left over from a previously loaded profile)
//   - a profile was unloaded (e.g. `swy off` elsewhere) → unset it
//   - state cleared without any loaded profile → only record the generation
//
// Every non-nil result ends with an assignment of SWITCHY_APPLIED_GEN so the
// hook does not repeat the work on the next prompt.
func ApplyScript(s Sync) []string {
	if s.Generation <= 0 || s.Generation <= s.AppliedGeneration {
		return nil
	}

	var lines []string
	switch {
	case s.CurrentProfile != "" && s.CurrentVars != nil:
		if s.LoadedProfile != "" && s.LoadedVars != nil {
			lines = staleUnsetLines(s.LoadedVars, s.CurrentVars)
		}
		lines = append(lines, exportLines(s.CurrentVars)...)
		lines = append(lines, shell.FormatExport(ProfileEnv, s.CurrentProfile))
	case s.LoadedProfile != "":
		lines = unsetLines(s.LoadedVars)
	default:
		// State was cleared and this shell has nothing loaded; just mark it as caught up.
	}
	lines = append(lines, shell.FormatExport(AppliedGenEnv, strconv.Itoa(s.Generation)))
	return lines
}

// Off describes a shell session that should drop its loaded profile.
type Off struct {
	// LoadedProfile is the profile this shell last applied ("" when none).
	LoadedProfile string
	// LoadedVars holds the variables of LoadedProfile; nil when unknown or
	// the profile no longer exists.
	LoadedVars map[string]string
}

// OffScript returns the shell lines that remove the loaded profile's
// variables (and SWITCHY_PROFILE) from the current shell, or nil when the
// shell has no profile loaded.
func OffScript(o Off) []string {
	if o.LoadedProfile == "" {
		return nil
	}
	return unsetLines(o.LoadedVars)
}

// exportLines returns sorted `export KEY='value'` lines for vars.
func exportLines(vars map[string]string) []string {
	keys := sortedKeys(vars)
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, shell.FormatExport(k, vars[k]))
	}
	return lines
}

// unsetLines returns a single `unset KEY...` line (plus SWITCHY_PROFILE),
// or nil when there is nothing to unset.
func unsetLines(vars map[string]string) []string {
	names := sortedKeys(vars)
	names = append(names, ProfileEnv)
	return []string{"unset " + strings.Join(names, " ")}
}

// staleUnsetLines returns a single `unset KEY...` line for keys present in
// loaded but absent from current (SWITCHY_PROFILE excluded — it is
// re-exported), or nil when nothing is stale.
func staleUnsetLines(loaded, current map[string]string) []string {
	var stale []string
	for _, k := range sortedKeys(loaded) {
		if _, ok := current[k]; !ok && k != ProfileEnv {
			stale = append(stale, k)
		}
	}
	if len(stale) == 0 {
		return nil
	}
	return []string{"unset " + strings.Join(stale, " ")}
}

func sortedKeys(vars map[string]string) []string {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
