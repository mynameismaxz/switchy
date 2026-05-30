# Switchy (`swy`)

A terminal tool for managing and switching named environment variable profiles — no more manually editing `~/.zshrc` or `~/.bashrc`.

```
⚡ Switchy
Shell: zsh  •  Current: local

>  *  local
      staging
      prod

↑↓/jk navigate  a add  p dup  e edit  d delete  s use  x view  E export  I import  q quit
```

---

## Why

Switching environments by hand is slow and error-prone:

1. Open `~/.zshrc`
2. Find `ANTHROPIC_BASE_URL`
3. Edit the value
4. Run `source ~/.zshrc`

Switchy replaces that with one command.

---

## Install

### Homebrew (macOS and Linux)

```bash
brew tap mynameismaxz/tap
brew install switchy
```

Or in one line:

```bash
brew install mynameismaxz/tap/switchy
```

### From source

```bash
git clone https://github.com/mynameismaxz/switchy
cd switchy
make install
```

Requires Go 1.25+. The binary is installed to `$GOPATH/bin/swy`.

### Build only

```bash
make build
# binary at ./bin/swy
```

---

## Quick start

```bash
# Create a profile
swy set local ANTHROPIC_BASE_URL=http://localhost:8080 API_KEY=dev-key

# Select it — the apply command is copied to your clipboard automatically
swy use local
# → paste and press Enter to apply

# Or install the shell helper once (no clipboard needed)
swy init
swyuse local
```

---

## Commands

### `swy set <profile> KEY=VALUE [KEY=VALUE...]`

Create a new profile or update an existing one. Unspecified keys are left unchanged.

```bash
swy set local ANTHROPIC_BASE_URL=http://localhost:8080 DEBUG=true
swy set prod  ANTHROPIC_BASE_URL=https://api.anthropic.com API_KEY=prod-key
```

### `swy list`

List all profiles. The current profile is marked with `*`.

```
* local
  staging
  prod
```

### `swy show <profile>`

Display a profile's variables. Sensitive values (keys containing `KEY`, `TOKEN`, `SECRET`, `PASSWORD`) are masked by default.

```
Profile: local
  ANTHROPIC_BASE_URL             = http://localhost:8080
  API_KEY                        = ****
```

### `swy export <profile>`

Print shell-compatible export statements. Designed for use with `eval`.

```bash
eval "$(swy export local)"
```

Output:
```bash
export ANTHROPIC_BASE_URL='http://localhost:8080'
export API_KEY='dev-key'
export SWITCHY_PROFILE='local'
```

### `swy use <profile>`

Mark a profile as selected and copy the apply command to your clipboard automatically. Does not modify the rc file.

```bash
swy use local
# Profile "local" selected.
# Copied to clipboard — paste and press Enter to apply.
# To persist:    swy use local --persistent
```

Just paste and press Enter — no typing required. If no clipboard tool is available, the `eval` command is printed instead:

```bash
# To apply now:  eval "$(swy export local)"
```

> **Clipboard support:** macOS uses `pbcopy` (always available). Linux requires `xclip`, `xsel`, or `wl-copy`.

### `swy use <profile> --persistent`

Write the profile into a managed block in your shell rc file so future shells load it automatically.

```bash
swy use prod --persistent
source ~/.zshrc
```

Writes to `~/.zshrc` (zsh) or `~/.bashrc` (bash). A backup is created at `~/.zshrc.switchy.bak` before the first modification.

### `swy use <profile> --inline`

Edit existing `export KEY=...` lines in your rc file in place — no managed block, no `eval`. For each variable in the profile, Switchy updates the **first non-commented matching `export`** it finds, or appends the line at the end if none exists.

```bash
swy use local --inline
# source command is copied to clipboard automatically (fallback shown if unavailable)
```

Useful when you already have `export ANTHROPIC_BASE_URL=...` written by hand somewhere in your rc file and just want its value swapped, without introducing a Switchy-managed block. Commented-out `# export ...` lines are never touched. A backup is still created on the first write.

> Choose `--persistent` if you prefer all Switchy-managed exports grouped in a clearly delimited block; choose `--inline` if you'd rather Switchy edit the lines you already have.

### `swy current`

Show the last profile selected through Switchy.

```
Current profile:  local
Activation mode:  session
To use now:       eval "$(swy export local)"
To persist:       swy use local --persistent
```

### `swy delete <profile>`

Delete a profile. If the profile is currently selected, `--force` is required.

```bash
swy delete staging
swy delete local --force
```

### `swy export-config [output-file]`

Export all profiles to a portable JSON file for backup or transfer.

```bash
# Print to stdout
swy export-config

# Save to a file
swy export-config my-profiles.json
swy export-config -o my-profiles.json
```

### `swy import-config <input-file>`

Import profiles from a JSON file created by `swy export-config`. By default, profiles are merged (existing profiles with the same name are updated). Use `--replace` to remove all existing profiles first.

```bash
# Merge profiles from a file
swy import-config my-profiles.json

# Replace all existing profiles
swy import-config my-profiles.json --replace
```

### `swy completion <shell>`

Generate shell completion script for autocomplete support. Profile names are dynamically completed for `swy use`.

```bash
# Bash
source <(swy completion bash)
swy completion bash >> ~/.bashrc

# Zsh
source <(swy completion zsh)
swy completion zsh >> ~/.zshrc

# Fish
swy completion fish > ~/.config/fish/completions/swy.fish
```

Once enabled, `swy use <TAB>` will show available profile names with the current profile marked.

### `swy init`

Install the `swyuse()` helper function into your rc file (idempotent — safe to run multiple times).

```bash
swy init
# then in any shell:
swyuse local
```

The installed function:

```bash
swyuse() {
  eval "$(swy export "$1")"
}
```

### `swy` (no arguments)

Launch the interactive TUI.

---

## TUI

Run `swy` with no arguments to open the interactive terminal UI.

```
⚡ Switchy
Shell: zsh  •  Current: local

>  *  local
      staging
      prod

↑↓/jk navigate  a add  p dup  e edit  d delete  s use  x view  E export  I import  q quit
```

**Keybindings:**

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `a` | Add new profile |
| `p` | Duplicate selected profile |
| `e` | Edit selected profile |
| `d` | Delete selected profile |
| `s` | Select profile (session mode) |
| `x` | Show export commands for profile |
| `E` | Export all profiles to a JSON file |
| `I` | Import profiles from a JSON file |
| `q` | Quit |

---

## Shell detection

Switchy detects your shell from `$SHELL`:

| Shell | RC file |
|-------|---------|
| zsh | `~/.zshrc` |
| bash | `~/.bashrc` |

Override with `--shell`:

```bash
swy use local --persistent --shell bash
swy init --shell zsh
```

---

## How rc file modification works

Switchy only ever writes to a **managed block** — it never touches anything outside it:

```bash
# >>> switchy start >>>
export ANTHROPIC_BASE_URL='https://api.anthropic.com'
export API_KEY='prod-key'
export SWITCHY_PROFILE='prod'
# <<< switchy end <<<
```

The block is replaced in-place on each `swy use --persistent`. A backup (`~/.zshrc.switchy.bak`) is created before the first write.

---

## Storage

Configuration is stored in `~/.config/switchy/`:

```
~/.config/switchy/
├── profiles.json   # all profiles and their variables
└── state.json      # currently selected profile and activation mode
```

Files are created with `0600` permissions. Values are stored in plaintext — Switchy is not a secret manager.

---

## Validation

- **Profile names:** `^[A-Za-z0-9_-]+$`
- **Env variable keys:** `^[A-Z_][A-Z0-9_]*$` (uppercase only)
- **Values:** any string except multiline

---

## Development

```bash
make build    # build to ./bin/swy
make install  # install to $GOPATH/bin
make test     # run tests
make lint     # go vet
make clean    # remove ./bin
```

**Project layout:**

```
cmd/swy/            binary entrypoint
internal/
  config/           JSON storage (profiles.json, state.json)
  validate/         input validation
  profile/          profile CRUD
  shell/            shell detection and rc file management
  cmd/              cobra CLI commands
  tui/              bubbletea TUI
```

---

## License

MIT
