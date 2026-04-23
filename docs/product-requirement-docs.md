# Product Requirements Document (v2)
# Switchy (`swy`)

## 1. Document Information

- **Product Name:** Switchy
- **CLI Command:** `swy`
- **Version:** PRD v2
- **Status:** Draft
- **Primary Audience:** Product, Engineering, Design
- **Implementation Language:** Go

---

## 2. Executive Summary

Switchy is a terminal-based tool for managing and switching between named environment variable profiles, such as `local`, `staging`, and `production`, without requiring users to manually edit shell configuration files like `~/.zshrc` or `~/.bashrc`.

Switchy will provide:

- a **CLI** for scripting and fast terminal usage
- a **TUI** for easier interactive profile management
- automatic detection of the user’s shell (`zsh` or `bash`)
- safe management of shell environment activation
- profile storage in a local config directory

The product is intended to reduce friction, errors, and time spent switching environment variables like `ANTHROPIC_BASE_URL`, API keys, and related configuration values.

---

# 3. Problem Statement

Today, users often switch environments by:

1. opening `~/.zshrc` or `~/.bashrc`
2. locating a target variable such as `ANTHROPIC_BASE_URL`
3. manually editing the variable value
4. reloading the shell configuration with `source ~/.zshrc` or `source ~/.bashrc`

This workflow is:

- repetitive
- slow
- error-prone
- inconvenient for users who switch often

Users need a simpler way to define named environment setups and switch between them quickly from the terminal.

---

# 4. Goals

## 4.1 Primary Goals
- Make switching environment variable sets fast and easy
- Eliminate manual editing of shell rc files for normal workflows
- Support `zsh` and `bash`
- Provide a clean, intuitive CLI
- Provide an optional TUI for easier management
- Be distributed as a lightweight standalone Go binary

## 4.2 Secondary Goals
- Support both session-only and persistent environment activation
- Allow users to inspect, edit, and manage multiple profiles
- Be safe in how it modifies shell rc files
- Be understandable for both technical and less-technical users

---

# 5. Non-Goals

Switchy v1 will not:

- support shells beyond `zsh` and `bash`
- encrypt or securely store secrets
- integrate with external secret managers
- manage project-local `.env` files
- sync profiles across machines
- provide GUI desktop applications

---

# 6. Target Users

## 6.1 Primary Users
- developers switching between local, staging, and production services
- users who often modify environment variables in terminal workflows
- developers using `zsh` or `bash` on macOS or Linux

## 6.2 Secondary Users
- terminal users who prefer interactive UIs
- engineers who want a simpler way to manage API endpoints and credentials

---

# 7. User Stories

## 7.1 Core User Stories
1. As a developer, I want to save named profiles of environment variables so I can reuse them easily.
2. As a developer, I want to switch to a profile with one command instead of editing shell files manually.
3. As a developer, I want Switchy to detect whether I am using `zsh` or `bash` so that it behaves correctly.
4. As a developer, I want to list available profiles so I know what configurations I have.
5. As a developer, I want to view the last activated profile so I can confirm what I selected.

## 7.2 TUI User Stories
6. As a user, I want a terminal UI to create, edit, and delete profiles more easily.
7. As a user, I want to switch profiles from a menu so I do not need to remember commands.
8. As a user, I want to see my shell type and activation mode in the interface.

## 7.3 Advanced User Stories
9. As a developer, I want to activate a profile in my current shell session immediately.
10. As a developer, I want to activate a profile persistently for future shell sessions.
11. As a developer, I want safe rc-file updates that do not damage unrelated shell config.

---

# 8. Product Decisions

This section defines decisions that remove ambiguity.

## 8.1 Activation Modes
Switchy supports two environment activation modes:

### Session Mode
Applies environment variables to the current shell session by printing shell export statements.

Example:
```bash
eval "$(swy export local)"
```

### Persistent Mode
Writes a managed block into the appropriate shell rc file so that future shell sessions load the selected profile automatically.

Example:
```bash
swy use local --persistent
```

## 8.2 Default Recommended Activation Method
For safety and predictability, the **recommended default** is:

- **session mode** for immediate use in the current shell
- **persistent mode** only when explicitly requested

### Product rule
- `swy export <profile>` outputs shell exports
- `swy use <profile>` records the selected profile and prints next-step guidance
- `swy use <profile> --persistent` updates the shell rc file managed block

### Rationale
A CLI binary cannot directly mutate the parent shell environment. Session mode is the least surprising and safest approach.

---

# 9. Scope

## 9.1 MVP Scope
MVP includes:

- CLI written in Go
- support for `zsh` and `bash`
- named profile storage
- profile CRUD operations
- session-mode export
- persistent-mode activation with managed rc-file block
- current profile metadata tracking
- shell detection
- optional shell initialization command
- basic TUI is **not included in MVP**

## 9.2 Post-MVP Scope
Post-MVP may include:

- TUI support
- shell override improvements
- import/export profile files
- profile duplication
- secret masking improvements
- support for more shells

---

# 10. Functional Requirements

## 10.1 Profile Model
A profile is a named map of environment variables.

Example:
```json
{
  "local": {
    "ANTHROPIC_BASE_URL": "http://localhost:8080",
    "API_KEY": "dev-key"
  },
  "prod": {
    "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
    "API_KEY": "prod-key"
  }
}
```

## 10.2 Profile Management
Switchy must support:

- create profile
- update profile
- delete profile
- list profiles
- inspect profile details

## 10.3 Command Set
Required commands for v1:

```bash
swy set <profile> KEY=VALUE [KEY=VALUE...]
swy list
swy show <profile>
swy export <profile>
swy use <profile>
swy use <profile> --persistent
swy current
swy delete <profile>
swy init
```

Optional later:
```bash
swy            # launch TUI
```

---

# 11. Command Semantics

## 11.1 `swy set <profile> KEY=VALUE [KEY=VALUE...]`
Creates a new profile if it does not exist.  
Updates the profile if it already exists.

### Rules
- specified keys are added or updated
- unspecified existing keys remain unchanged
- invalid key names must be rejected
- malformed `KEY=VALUE` input must be rejected

Example:
```bash
swy set local ANTHROPIC_BASE_URL=http://localhost:8080 DEBUG=true
```

## 11.2 `swy list`
Lists all profile names.

### Default output
- profile names only
- current/last-selected profile should be marked if known

Example:
```bash
* local
  staging
  prod
```

## 11.3 `swy show <profile>`
Displays variables for a profile.

### Rules
- likely-sensitive values may be masked by default
- an optional future flag may reveal full values

## 11.4 `swy export <profile>`
Prints shell-compatible export statements for the selected profile.

### Example output
```bash
export ANTHROPIC_BASE_URL='http://localhost:8080'
export API_KEY='dev-key'
export SWITCHY_PROFILE='local'
```

### Rules
- output must be valid for `bash` and `zsh`
- values must be safely quoted
- output is intended for:
```bash
eval "$(swy export local)"
```

## 11.5 `swy use <profile>`
Marks the profile as the currently selected profile in Switchy state and prints guidance for activation.

### Behavior
- validate profile exists
- write `currentProfile` into state metadata
- print next-step guidance
- does **not** modify rc file unless `--persistent` is provided

## 11.6 `swy use <profile> --persistent`
Activates the profile persistently.

### Behavior
- detect shell
- determine target rc file
- update managed Switchy block in that file
- write state metadata
- print success and reload guidance

## 11.7 `swy current`
Shows the last profile selected through Switchy.

### Important definition
`current` refers to the last profile recorded by Switchy in local metadata.  
It does **not** guarantee that the current terminal session has those variables loaded.

## 11.8 `swy delete <profile>`
Deletes a profile.

### Rules
- deleting a missing profile must fail clearly
- deleting the current profile must require confirmation or a force flag in non-interactive mode
- if deleted profile is stored as `currentProfile`, current state becomes unset

## 11.9 `swy init`
Installs shell helper integration into the supported rc file.

Example inserted function:
```bash
swyuse() {
  eval "$(swy export "$1")"
}
```

### Rules
- use a managed block
- do not duplicate installation if already present
- print what was installed and where

---

# 12. Shell Detection Requirements

## 12.1 Supported Shells
- `zsh`
- `bash`

## 12.2 Detection Strategy
Primary detection uses `$SHELL`.

### Mapping
- `*zsh*` → `~/.zshrc`
- `*bash*` → `~/.bashrc`

## 12.3 Manual Override
Switchy should support manual shell override.

Example:
```bash
swy use local --persistent --shell zsh
```

## 12.4 Unsupported Shell Behavior
If shell detection fails or shell is unsupported:
- return a non-zero exit code
- print a clear message
- suggest `--shell zsh` or `--shell bash` if appropriate

---

# 13. RC File Modification Strategy

This is a critical safety requirement.

## 13.1 Managed Block Requirement
Switchy must only write to a dedicated managed block in the shell rc file.

Example:
```bash
# >>> switchy start >>>
export ANTHROPIC_BASE_URL='http://localhost:8080'
export API_KEY='dev-key'
export SWITCHY_PROFILE='local'
# <<< switchy end <<<
```

## 13.2 Update Rules
- If the managed block exists, replace the full block
- If the block does not exist, append it to the end of the file
- Do not modify unrelated content outside the block
- Preserve user formatting and content outside the block

## 13.3 Backup and Write Safety
- On first persistent write, create a backup file:
  - `~/.zshrc.switchy.bak` or `~/.bashrc.switchy.bak`
- Write changes atomically where practical
- If write fails, original file must remain unchanged

---

# 14. Current Profile Definition

## 14.1 Source of Truth
Switchy must track current state in a metadata file.

Example:
```json
{
  "currentProfile": "local",
  "lastActivationMode": "persistent"
}
```

## 14.2 Meaning of “Current”
For Switchy, “current profile” means:
- the last profile selected through Switchy and recorded in state metadata

It does not necessarily mean:
- the environment of every currently open shell tab
- the env actually loaded into a specific shell session

## 14.3 Rationale
Because shell sessions are independent, Switchy cannot reliably introspect all active terminal sessions.

---

# 15. Data Storage Requirements

## 15.1 Config Directory
Store configuration in:

```bash
~/.config/switchy/
```

## 15.2 Files
### Profiles
```bash
~/.config/switchy/profiles.json
```

### State
```bash
~/.config/switchy/state.json
```

## 15.3 File Permissions
- configuration files should be created with restrictive permissions where practical
- recommended file mode for sensitive content: `0600`

## 15.4 Config Versioning
Files should include a version field to support future migrations.

Example:
```json
{
  "version": 1,
  "profiles": {
    "local": {
      "ANTHROPIC_BASE_URL": "http://localhost:8080"
    }
  }
}
```

---

# 16. Input Validation Rules

## 16.1 Profile Names
Allowed characters:
- letters
- numbers
- `_`
- `-`

### Rules
- profile name cannot be empty
- profile names are case-sensitive
- spaces are not allowed in MVP

Recommended regex:
```regex
^[A-Za-z0-9_-]+$
```

## 16.2 Environment Variable Names
MVP env keys must follow standard shell env naming.

Recommended regex:
```regex
^[A-Z_][A-Z0-9_]*$
```

### Rules
- lowercase keys are rejected in MVP
- invalid names must return a clear error

## 16.3 Environment Variable Values
- empty values are allowed
- values may contain spaces and special characters
- multiline values are not supported in MVP unless safe quoting is explicitly implemented

---

# 17. Quoting and Escaping Requirements

Switchy must produce shell-safe exports compatible with both `bash` and `zsh`.

## 17.1 Export Format
Preferred export format:
```bash
export KEY='value'
```

## 17.2 Escaping
Values must be correctly escaped for:
- single quotes
- spaces
- dollar signs
- ampersands
- backslashes

## 17.3 Unsupported Value Cases
If a value cannot be safely represented in MVP:
- return a clear error
- explain the unsupported format

---

# 18. Security and Privacy Requirements

## 18.1 Plaintext Storage
Switchy stores profile values locally in plaintext in MVP.

## 18.2 No Secret-Manager Claim
Switchy is not a secure secret-management product in v1.

## 18.3 Sensitive Output Handling
Switchy should avoid exposing secrets unnecessarily.

### Rules
- `swy list` shows profile names only
- `swy show <profile>` may mask likely-sensitive values by default
- likely-sensitive keys include names containing:
  - `KEY`
  - `TOKEN`
  - `SECRET`
  - `PASSWORD`

## 18.4 User Responsibility
Users are responsible for machine-level security and file-system access controls.

---

# 19. Error Handling Requirements

All failures must:
- return non-zero exit codes
- print concise, actionable messages

## 19.1 Required Error Cases
Switchy must handle:

- profile not found
- invalid profile name
- invalid env key
- malformed `KEY=VALUE`
- unsupported shell
- shell not detectable
- rc file not writable
- config file malformed
- config directory cannot be created
- duplicate init installation attempt
- persistent activation failure

## 19.2 Error Message Quality
Messages should:
- identify the problem
- state what failed
- suggest what user can do next if possible

Example:
```bash
Error: unsupported shell "/bin/fish".
Supported shells: zsh, bash.
Try: swy use local --persistent --shell zsh
```

---

# 20. First-Run Experience

## 20.1 Initial Setup
On first run, if the config directory does not exist:
- create it automatically

## 20.2 Empty State
If no profiles exist:
- `swy list` should print a helpful empty-state message
- future TUI should show “No profiles yet”

Example:
```bash
No profiles found.
Create one with:
  swy set local ANTHROPIC_BASE_URL=http://localhost:8080
```

## 20.3 Init Guidance
If user tries to use session-mode helper flow before `swy init`, Switchy should still work via `export`, and may suggest:
```bash
Run `swy init` to install the `swyuse` shell helper.
```

---

# 21. TUI Requirements

## 21.1 TUI Release Scope
TUI is **not part of MVP**.  
TUI is targeted for **v1.1 or later**.

## 21.2 Intended Launch Command
```bash
swy
```

## 21.3 Planned TUI Features
- profile list
- profile detail view
- add/edit/delete profile
- activate profile
- show current shell
- show current selected profile
- keyboard-driven navigation

## 21.4 Suggested Keybindings
- `a` = add
- `e` = edit
- `d` = delete
- `s` = switch
- `x` = export
- `q` = quit

---

# 22. Non-Functional Requirements

## 22.1 Performance
- common commands should complete quickly for small-to-medium profile sets
- startup should feel near-instant for normal usage

## 22.2 Reliability
- rc file updates must preserve unrelated content
- backup must be created before first persistent modification
- malformed state/config should fail safely

## 22.3 Portability
- support macOS and Linux in v1
- distributed as a standalone Go binary

## 22.4 Offline Usage
- all core functionality must work offline

---

# 23. Technical Requirements

## 23.1 Language
- Go

## 23.2 Suggested Libraries
### CLI
- standard library or `cobra`

### TUI
- `bubbletea`
- `bubbles`
- `lipgloss`

## 23.3 Shell Compatibility
- exported statements must be valid in both `bash` and `zsh`

---

# 24. Acceptance Criteria

## 24.1 CLI Happy Path
- user can create a profile
- user can update a profile
- user can list profiles
- user can inspect a profile
- user can export a profile to shell-compatible output
- user can mark a profile as current
- user can activate a profile persistently
- Switchy detects `zsh` or `bash` correctly
- Switchy updates the correct rc file managed block
- Switchy leaves unrelated rc-file content untouched
- user can delete a profile

## 24.2 Edge Case Acceptance
- values with spaces are exported safely
- invalid env key names are rejected
- invalid shell type returns clear error
- malformed config file returns clear error
- deleting current profile clears current state
- managed block replacement does not duplicate content
- unsupported shell suggests manual override

## 24.3 Init Acceptance
- `swy init` installs shell helper exactly once
- helper is inserted into a managed block
- repeated init does not duplicate the block

---

# 25. Example User Flows

## 25.1 Create and Export to Current Shell
```bash
swy set local ANTHROPIC_BASE_URL=http://localhost:8080
eval "$(swy export local)"
```

## 25.2 Set Current Profile Without Persistent Activation
```bash
swy use local
swy current
```

Expected output:
```bash
Current profile: local
Activation mode: none-applied
To use now: eval "$(swy export local)"
To persist: swy use local --persistent
```

## 25.3 Persistent Activation
```bash
swy use prod --persistent
source ~/.zshrc
```

## 25.4 Shell Helper Installation
```bash
swy init
swyuse local
```

---

# 26. Risks and Mitigations

## 26.1 Parent Shell Limitation
A child process cannot directly modify the parent shell’s environment.

### Mitigation
- support `export` output
- support `init` helper
- document behavior clearly

## 26.2 RC File Corruption Risk
Incorrect rc-file editing could break user shell startup.

### Mitigation
- use managed block only
- create backups
- write atomically where possible
- never edit outside managed block

## 26.3 Secret Exposure Risk
Stored profile values may contain secrets.

### Mitigation
- state plaintext storage explicitly
- mask sensitive values in common displays
- use restrictive file permissions where practical

---

# 27. Open Questions

These items still require a final product decision:

1. Should `swy use <profile>` default to session guidance only, or also print a one-line `eval` command suggestion?
2. Should `swy show <profile>` mask sensitive values by default or reveal them by default?
3. Should deleting the current profile require interactive confirmation only, or also support `--force`?
4. Should multiline env values be explicitly unsupported in v1, or allowed with stricter escaping?
5. Should `swy export` always include `SWITCHY_PROFILE` metadata?

---

# 28. Recommended Delivery Plan

## Phase 1: CLI MVP
- config and state storage
- profile CRUD
- shell detection
- export command
- current command
- persistent activation via managed block
- init command

## Phase 2: Hardening
- masking sensitive values
- better error messages
- validation improvements
- backup/restore reliability
- test coverage for edge cases

## Phase 3: TUI
- profile list
- add/edit/delete flow
- activation flow
- shell status display

---

# 29. Success Metrics

Switchy is successful if:

- users can switch environment profiles in seconds
- users no longer need to manually edit shell rc files for normal workflows
- profile switching is understandable and low-error
- shell startup files remain safe and intact after use
- users can manage multiple profiles with minimal command memorization

---

# 30. Appendix: Example Managed Blocks

## 30.1 RC Activation Block
```bash
# >>> switchy start >>>
export ANTHROPIC_BASE_URL='http://localhost:8080'
export API_KEY='dev-key'
export SWITCHY_PROFILE='local'
# <<< switchy end <<<
```

## 30.2 Init Helper Block
```bash
# >>> switchy init start >>>
swyuse() {
  eval "$(swy export "$1")"
}
# <<< switchy init end <<<
```
