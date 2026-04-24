# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- `make build` — build to `./bin/swy` (injects `version`/`commit`/`date` via `-ldflags` into `cmd/swy/main.go`)
- `make install` — `go install` to `$GOPATH/bin/swy`
- `make test` — `go test ./...`
- `make lint` — `go vet ./...`
- Run a single test: `go test ./internal/shell -run TestUpsertExports -v`
- Requires Go **1.25+** (see `go.mod`).

## Architecture

`swy` is a Cobra CLI + Bubble Tea TUI for managing named env-var profiles and applying them to the user's shell. The flow is `cmd/swy/main.go` → `internal/cmd.NewRootCmd` → subcommand or TUI.

### Layered packages (strict downward dependencies)

```
cmd/  →  tui/, profile/, shell/, clipboard/, validate/, config/
profile/  →  config/
shell/  →  (self-contained, file I/O only)
config/  →  (self-contained, JSON + paths)
validate/  →  (pure functions)
```

`cmd/` is the only layer that orchestrates across packages. `tui/` is an alternative front-end that calls the same `profile`/`shell`/`config` APIs as the CLI subcommands — keep parity when adding behavior to one.

### Storage (`internal/config`)

Two JSON files under `~/.config/switchy/` (mode `0600`, written atomically via temp+rename):
- `profiles.json` — `{version, profiles: {name: {KEY: VALUE}}}`
- `state.json` — `{currentProfile, lastActivationMode}` where mode is `"session"` | `"persistent"` | `""`

Tests inject a temp dir via the package-private `configDirOverride` in `paths.go`.

### Three ways `swy use` applies a profile

`internal/cmd/use.go` branches on flags — when adding a new mode, mirror all three:

1. **Session (default)** — saves state, copies `eval "$(swy export <name>)"` to clipboard via `internal/clipboard`. Falls back to printing the command if no clipboard tool is found (`pbcopy` on macOS; `xclip`/`xsel`/`wl-copy` on Linux).
2. **`--persistent`** — `shell.WriteActivationBlock` writes a managed block delimited by `# >>> switchy start >>>` / `# <<< switchy end <<<` into `~/.zshrc` or `~/.bashrc`. The block is replaced in-place; nothing outside it is touched.
3. **`--inline`** — `shell.UpsertExports` rewrites individual `export KEY=...` lines in the rc file (updating the first non-commented match per key, appending new keys at the end).

`swy init` writes a separate managed block (`# >>> switchy init start >>>`) containing the `swyuse()` helper. Both block writers call `BackupIfNeeded`, which creates `~/.zshrc.switchy.bak` once before the first modification.

### Shell escaping

All export emission goes through `shell.FormatExport` → `SingleQuoteEscape` (turns `'` into `'\''`). Never hand-format export lines elsewhere — the value comes from arbitrary user input.

### Validation rules (`internal/validate`)

- Profile names: `^[A-Za-z0-9_-]+$`
- Env keys: `^[A-Z_][A-Z0-9_]*$` (uppercase only, enforced on set)
- Values: any string except those containing `\n`
- `IsSensitiveKey` masks values whose key contains `KEY`/`TOKEN`/`SECRET`/`PASSWORD` in `swy show` output.

The selected profile is always also exported as `SWITCHY_PROFILE=<name>` — preserve this when adding new activation modes.
