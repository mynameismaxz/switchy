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
brew install mynameismaxz/tap/switchy
```

On Homebrew 6+, installing with the fully qualified name trusts only the Switchy formula. If you prefer to explicitly trust the entire [tap](https://github.com/mynameismaxz/homebrew-tap), run:

```bash
brew tap mynameismaxz/tap
brew trust mynameismaxz/tap
brew install mynameismaxz/tap/switchy
```

See [Homebrew's tap trust documentation](https://docs.brew.sh/Tap-Trust) for the difference between formula and tap trust.

### Linux packages

Download the package for your distribution and CPU architecture (`amd64`/`x86_64` or `arm64`/`aarch64`) from [GitHub Releases](https://github.com/mynameismaxz/switchy/releases). Run the matching command from the download directory, replacing `PACKAGE` with the downloaded filename without its extension:

```bash
# Debian / Ubuntu
sudo apt install ./PACKAGE.deb

# Fedora / RHEL
sudo dnf install ./PACKAGE.rpm
```

The package installs `swy` to `/usr/bin/swy`. To upgrade, download and install the newer package the same way.

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

## Update

### Homebrew

```bash
brew update
brew upgrade mynameismaxz/tap/switchy
```

If Homebrew reports that an existing installation is from an untrusted tap, trust just the Switchy formula and retry the upgrade:

```bash
brew trust --formula mynameismaxz/tap/switchy
brew upgrade mynameismaxz/tap/switchy
```

### From source

In your existing Switchy checkout, update `main` and reinstall:

```bash
git switch main && git pull --ff-only origin main && make install
```

If you run the binary directly from `./bin/swy`, use `make build` instead of `make install`.

### After updating

Run `swy version` to check the installed version. A source build reports `dev` by default. Your saved profiles stay in `~/.config/switchy/`.

If you use the shell hook, rerun `swy init` to pick up any hook changes. Then open a new terminal or reload your shell configuration:

```bash
swy init
source ~/.zshrc   # or: source ~/.bashrc
```

---

## Quick start

```bash
# One-time setup: install the shell hook into your rc file
swy init
source ~/.zshrc   # or: source ~/.bashrc

# Create a profile
swy set local ANTHROPIC_BASE_URL=http://localhost:8080 API_KEY=dev-key

# Apply it — takes effect in this shell immediately, no clipboard, no pasting
swy use local

# Switch anytime; other open shells follow along on their next prompt
swy use prod

# Unload the profile from this shell
swy off
```

Without the hook, `swy use` still works via the clipboard: the apply command is copied and you paste it to apply.

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

Mark a profile as selected and apply it.

With the **shell hook** installed (see `swy init`), the profile's variables are applied to the current shell immediately — no clipboard, no pasting, no `source`:

```bash
swy use local
# ⚡ applied "local" to this shell
```

Other shells with the hook installed also pick up the change on their next prompt, and new terminals start with the selected profile loaded automatically.

Without the hook, the apply command is copied to your clipboard instead:

```bash
swy use local
# Profile "local" selected.
# Copied to clipboard — paste and press Enter to apply.
# To persist:    swy use local --persistent
# Tip: run `swy init` once so `swy use` applies instantly — no clipboard needed.
```

Just paste and press Enter — no typing required. If no clipboard tool is available, the `eval` command is printed instead:

```bash
# To apply now:  eval "$(swy export local)"
```

> **Clipboard support:** macOS uses `pbcopy` (always available). Linux requires `xclip`, `xsel`, or `wl-copy`.

### `swy use <profile> --persistent`

Write the profile into a managed block in your shell rc file so future shells load it automatically. With the hook installed, it is also applied to the current shell right away — no `source` needed.

```bash
swy use prod --persistent
# Profile "prod" activated in ~/.zshrc.
# Applied to this shell — no source needed.
```

Writes to `~/.zshrc` (zsh) or `~/.bashrc` (bash). A backup is created at `~/.zshrc.switchy.bak` before the first modification.

### `swy use <profile> --inline`

Edit existing `export KEY=...` lines in your rc file in place — no managed block, no `eval`. For each variable in the profile, Switchy updates the **first non-commented matching `export`** it finds, or appends the line at the end if none exists.

```bash
swy use local --inline
source ~/.zshrc
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

Install the swy shell hook into your rc file (idempotent — safe to run multiple times). This is the recommended setup: after sourcing your rc file once, switching is instant.

```bash
swy init
source ~/.zshrc   # once — after this, switchy applies instantly
swy use local     # applied to this shell immediately
```

The installed managed block defines:

- **`swy()` wrapper** — applies profile changes from the TUI (`swy` with no arguments), `swy use`, and `swy off` in the current shell right away.
- **`swyuse()` helper** — kept for muscle memory: `swyuse local` behaves like `swy use local`.
- **A prompt hook** — on every prompt, the shell evals `swy hook-apply`, which applies the selected profile only if its state changed since the shell last synced. This is what keeps other terminals in sync.

```bash
swy() {
  case "${1-}" in
    "")  SWITCHY_HOOK=1 command swy "$@" && eval "$(command swy hook-apply)" ;;
    use) shift; SWITCHY_HOOK=1 command swy use "$@" && eval "$(command swy hook-apply)" ;;
    off) shift; eval "$(command swy off "$@")" ;;
    *)   command swy "$@" ;;
  esac
}
```

`swy init` upgrades older managed blocks in place, including hooks installed before TUI support. Source your rc file again or open a new terminal to load the updated hook.

### `swy hook <zsh|bash>`

Print the hook script for manual installation instead of using `swy init`:

```bash
swy hook zsh >> ~/.zshrc
swy hook bash >> ~/.bashrc
```

### `swy off`

Remove the active profile from the current shell (unset its variables, including `SWITCHY_PROFILE`) and clear the selection. With the hook installed, plain `swy off` does this for you; hooked shells in other terminals unload on their next prompt.

```bash
swy off
```

If a persistent block remains in your rc file, Switchy prints a note — new shells would still load it. Remove the block too with:

```bash
swy off --persistent
```

### `swy hook-apply`

Internal command used by the shell hook (hidden from help). Compares the state generation with the shell's `SWITCHY_APPLIED_GEN` marker and prints the exports needed to catch up. Useful for debugging the hook:

```bash
eval "$(swy hook-apply)"
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
| `s` | Use selected profile (session mode) |
| `x` | Show export commands for profile |
| `E` | Export all profiles to a JSON file |
| `I` | Import profiles from a JSON file |
| `q` | Quit |

With the shell hook installed, select a profile and press `s` to close the TUI and apply the profile immediately in the same shell. Other hooked shells apply it on their next prompt. Without the hook, Switchy copies the apply command to the clipboard or prints it if the clipboard is unavailable.

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
└── state.json      # current profile, activation mode, and generation counter
```

`state.json` also stores a `generation` counter that increases on every `swy use` / `swy off`. The shell hook compares it with each shell's `SWITCHY_APPLIED_GEN` marker to decide whether the shell needs to re-apply the profile — this is what makes switching sync across terminals.

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
  activation/       pure rendering of hook apply/off scripts
  config/           JSON storage (profiles.json, state.json)
  validate/         input validation
  profile/          profile CRUD
  shell/            shell detection, rc file management, hook script
  cmd/              cobra CLI commands
  tui/              bubbletea TUI
```

---

## License

MIT
