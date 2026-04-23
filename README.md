# Switchy (`swy`)

A terminal tool for managing and switching named environment variable profiles — no more manually editing `~/.zshrc` or `~/.bashrc`.

```
⚡ Switchy
Shell: zsh  •  Current: local

>  *  local
      staging
      prod

↑↓/jk navigate  a add  e edit  d delete  s switch  x export  q quit
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

# Apply it to your current shell session
eval "$(swy export local)"

# Or install the shell helper once (easier)
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

Mark a profile as selected and print activation guidance. Does not modify the rc file.

```bash
swy use local
# Profile "local" selected.
# To apply now:  eval "$(swy export local)"
# To persist:    swy use local --persistent
```

### `swy use <profile> --persistent`

Write the profile into a managed block in your shell rc file so future shells load it automatically.

```bash
swy use prod --persistent
source ~/.zshrc
```

Writes to `~/.zshrc` (zsh) or `~/.bashrc` (bash). A backup is created at `~/.zshrc.switchy.bak` before the first modification.

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

↑↓/jk navigate  a add  e edit  d delete  s switch  x export  q quit
```

**Keybindings:**

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `a` | Add new profile |
| `e` | Edit selected profile |
| `d` | Delete selected profile |
| `s` | Select profile (session mode) |
| `x` | Show export commands |
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
