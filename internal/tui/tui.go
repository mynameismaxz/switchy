package tui

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mynameismaxz/switchy/internal/clipboard"
	"github.com/mynameismaxz/switchy/internal/config"
	"github.com/mynameismaxz/switchy/internal/profile"
	"github.com/mynameismaxz/switchy/internal/shell"
	"github.com/mynameismaxz/switchy/internal/validate"
)

// ── Styles ───────────────────────────────────────────────────────────────────

var (
	purple = lipgloss.Color("#7C3AED")
	green  = lipgloss.Color("#10B981")
	red    = lipgloss.Color("#EF4444")
	gray   = lipgloss.Color("#6B7280")
	light  = lipgloss.Color("#9CA3AF")

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(purple)
	selectedStyle = lipgloss.NewStyle().Foreground(purple).Bold(true)
	currentStyle  = lipgloss.NewStyle().Foreground(green).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(gray)
	helpStyle     = lipgloss.NewStyle().Foreground(light)
	errorStyle    = lipgloss.NewStyle().Foreground(red).Bold(true)
	successStyle  = lipgloss.NewStyle().Foreground(green)
	labelStyle    = lipgloss.NewStyle().Foreground(purple).Bold(true)
)

// ── View states ──────────────────────────────────────────────────────────────

type viewState int

const (
	viewList          viewState = iota
	viewAddName                 // input profile name (add or duplicate)
	viewVarList                 // navigable list of pending vars
	viewDeleteConfirm           // delete confirmation
	viewMessage                 // success/error message
	viewEditOneVar              // text input for a single KEY=VALUE
	viewExportPath              // text input for export file path
	viewImportPath              // text input for import file path
)

// ── Model ────────────────────────────────────────────────────────────────────

type model struct {
	state    viewState
	profiles []string
	current  string
	shell    string
	cursor   int

	input       textinput.Model
	editTarget  string
	pendingVars map[string]string
	isAdding    bool
	formErr     string

	// duplicate flow
	duplicateSource string

	// import flow
	importReplace bool

	// var-list navigation
	varCursor  int
	varKeys    []string
	editingKey string

	message string
	isError bool

	// printed to stdout by Run() after the TUI exits
	finalMsg string

	width  int
	height int
}

func initialModel() model {
	ti := textinput.New()
	ti.CharLimit = 500

	m := model{
		state:       viewList,
		pendingVars: make(map[string]string),
		input:       ti,
	}
	m.reload()
	return m
}

func (m *model) reload() {
	names, cur, _ := profile.ListProfiles()
	m.profiles = names
	m.current = cur
	if sh, err := shell.Detect(""); err == nil {
		m.shell = string(sh.Type)
	} else {
		m.shell = "unknown"
	}
	if len(m.profiles) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.profiles) {
		m.cursor = len(m.profiles) - 1
	}
}

// ── Entry point ──────────────────────────────────────────────────────────────

func Run() error {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}
	if fm, ok := final.(model); ok && fm.finalMsg != "" {
		fmt.Println(fm.finalMsg)
	}
	return nil
}

// ── Bubbletea interface ──────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.state {
		case viewList:
			return m.updateList(msg)
		case viewAddName:
			return m.updateAddName(msg)
		case viewVarList:
			return m.updateVarList(msg)
		case viewEditOneVar:
			return m.updateEditOneVar(msg)
		case viewDeleteConfirm:
			return m.updateDeleteConfirm(msg)
		case viewExportPath:
			return m.updateExportPath(msg)
		case viewImportPath:
			return m.updateImportPath(msg)
		case viewMessage:
			m.state = viewList
			m.message = ""
			return m, nil
		}
	}
	return m, nil
}

// ── Update handlers ──────────────────────────────────────────────────────────

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.profiles)-1 {
			m.cursor++
		}

	case "a":
		m.state = viewAddName
		m.input.Reset()
		m.input.Placeholder = "my-profile"
		m.input.Focus()
		m.pendingVars = make(map[string]string)
		m.varKeys = []string{}
		m.varCursor = 0
		m.isAdding = true
		m.editTarget = ""
		m.duplicateSource = ""
		m.formErr = ""
		return m, textinput.Blink

	case "p":
		if len(m.profiles) == 0 {
			return m, nil
		}
		m.duplicateSource = m.profiles[m.cursor]
		m.state = viewAddName
		m.input.Reset()
		m.input.Placeholder = "my-profile-copy"
		m.input.Focus()
		m.pendingVars = make(map[string]string)
		m.varKeys = []string{}
		m.varCursor = 0
		m.isAdding = true
		m.editTarget = ""
		m.formErr = ""
		return m, textinput.Blink

	case "e":
		if len(m.profiles) == 0 {
			return m, nil
		}
		name := m.profiles[m.cursor]
		vars, err := profile.GetProfile(name)
		if err != nil {
			m.message = err.Error()
			m.isError = true
			m.state = viewMessage
			return m, nil
		}
		m.editTarget = name
		m.isAdding = false
		m.pendingVars = make(map[string]string, len(vars))
		maps.Copy(m.pendingVars, vars)
		m.varKeys = sortedKeys(m.pendingVars)
		m.varCursor = 0
		m.state = viewVarList
		m.formErr = ""
		return m, nil

	case "d":
		if len(m.profiles) == 0 {
			return m, nil
		}
		m.editTarget = m.profiles[m.cursor]
		m.state = viewDeleteConfirm

	case "s":
		if len(m.profiles) == 0 {
			return m, nil
		}
		name := m.profiles[m.cursor]
		evalCmd := fmt.Sprintf(`eval "$(swy export %s)"`, name)
		copied := clipboard.Copy(evalCmd)
		_ = config.SaveState(&config.StateFile{
			CurrentProfile:     name,
			LastActivationMode: "session",
		})
		if copied {
			m.finalMsg = fmt.Sprintf("Switched to %q — command copied to clipboard.\nPaste and press Enter to apply.", name)
		} else {
			m.finalMsg = fmt.Sprintf("Switched to %q.\nTo apply:  %s", name, evalCmd)
		}
		return m, tea.Quit

	case "x":
		if len(m.profiles) == 0 {
			return m, nil
		}
		name := m.profiles[m.cursor]
		vars, err := profile.GetProfile(name)
		if err != nil {
			m.message = "Error: " + err.Error()
			m.isError = true
			m.state = viewMessage
			return m, nil
		}
		keys := sortedKeys(vars)
		lines := make([]string, 0, len(keys)+1)
		for _, k := range keys {
			lines = append(lines, shell.FormatExport(k, vars[k]))
		}
		lines = append(lines, shell.FormatExport("SWITCHY_PROFILE", name))
		m.message = fmt.Sprintf(
			"Export for %q:\n\n%s\n\nTo activate:\n  eval \"$(swy export %s)\"",
			name, strings.Join(lines, "\n"), name,
		)
		m.isError = false
		m.state = viewMessage

	case "E":
		m.state = viewExportPath
		m.input.Reset()
		m.input.Placeholder = "switchy-profiles.json"
		m.input.Focus()
		m.formErr = ""
		return m, textinput.Blink

	case "I":
		m.state = viewImportPath
		m.input.Reset()
		m.input.Placeholder = "switchy-profiles.json"
		m.input.Focus()
		m.importReplace = false
		m.formErr = ""
		return m, textinput.Blink
	}
	return m, nil
}

func (m model) updateAddName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.duplicateSource = ""
		m.state = viewList
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.input.Value())
		if err := validate.ProfileName(name); err != nil {
			m.formErr = err.Error()
			return m, textinput.Blink
		}
		m.editTarget = name
		m.pendingVars = make(map[string]string)
		if m.duplicateSource != "" {
			if src, err := profile.GetProfile(m.duplicateSource); err == nil {
				maps.Copy(m.pendingVars, src)
			}
			m.duplicateSource = ""
		}
		m.varKeys = sortedKeys(m.pendingVars)
		m.varCursor = 0
		m.state = viewVarList
		m.input.Reset()
		m.formErr = ""
		return m, nil
	default:
		m.formErr = ""
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateVarList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = viewList
		m.reload()
		return m, nil

	case "up", "k":
		if m.varCursor > 0 {
			m.varCursor--
		}

	case "down", "j":
		if m.varCursor < len(m.varKeys)-1 {
			m.varCursor++
		}

	case "a":
		m.editingKey = ""
		m.input.Reset()
		m.input.Placeholder = "KEY=VALUE"
		m.input.Focus()
		m.formErr = ""
		m.state = viewEditOneVar
		return m, textinput.Blink

	case "enter":
		if len(m.varKeys) == 0 {
			m.formErr = "press 'a' to add a variable"
			return m, nil
		}
		key := m.varKeys[m.varCursor]
		m.editingKey = key
		m.input.Reset()
		m.input.SetValue(key + "=" + m.pendingVars[key])
		m.input.Focus()
		m.formErr = ""
		m.state = viewEditOneVar
		return m, textinput.Blink

	case "d":
		if len(m.varKeys) == 0 {
			return m, nil
		}
		key := m.varKeys[m.varCursor]
		delete(m.pendingVars, key)
		m.varKeys = sortedKeys(m.pendingVars)
		if m.varCursor >= len(m.varKeys) && m.varCursor > 0 {
			m.varCursor = len(m.varKeys) - 1
		}
		m.formErr = ""

	case "s":
		if m.isAdding && len(m.pendingVars) == 0 {
			m.formErr = "at least one KEY=VALUE is required"
			return m, nil
		}
		if err := profile.UpsertProfile(m.editTarget, m.pendingVars); err != nil {
			m.message = "Error: " + err.Error()
			m.isError = true
		} else {
			action := "updated"
			if m.isAdding {
				action = "created"
			}
			m.message = fmt.Sprintf("Profile %q %s.", m.editTarget, action)
			m.isError = false
			m.reload()
		}
		m.state = viewMessage
		return m, nil
	}
	return m, nil
}

func (m model) updateEditOneVar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = viewVarList
		m.editingKey = ""
		m.input.Reset()
		m.formErr = ""
		return m, nil
	case "enter":
		val := strings.TrimSpace(m.input.Value())
		if val == "" {
			m.formErr = "enter KEY=VALUE"
			return m, textinput.Blink
		}
		pairs, err := validate.ParseKVPairs([]string{val})
		if err != nil {
			m.formErr = err.Error()
			m.input.Reset()
			return m, textinput.Blink
		}
		// if the key was renamed, remove the old entry
		for newKey := range pairs {
			if m.editingKey != "" && m.editingKey != newKey {
				delete(m.pendingVars, m.editingKey)
			}
			m.pendingVars[newKey] = pairs[newKey]
		}
		m.varKeys = sortedKeys(m.pendingVars)
		// keep cursor on the edited/new key
		for newKey := range pairs {
			for i, k := range m.varKeys {
				if k == newKey {
					m.varCursor = i
					break
				}
			}
		}
		m.editingKey = ""
		m.input.Reset()
		m.formErr = ""
		m.state = viewVarList
		return m, nil
	default:
		m.formErr = ""
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if err := profile.DeleteProfile(m.editTarget, true); err != nil {
			m.message = "Error: " + err.Error()
			m.isError = true
		} else {
			m.message = fmt.Sprintf("Profile %q deleted.", m.editTarget)
			m.isError = false
			m.reload()
		}
		m.state = viewMessage
	case "n", "N", "esc":
		m.state = viewList
	}
	return m, nil
}

func (m model) updateExportPath(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = viewList
		return m, nil
	case "enter":
		path := strings.TrimSpace(m.input.Value())
		if path == "" {
			m.formErr = "enter a file path"
			return m, textinput.Blink
		}
		pf, err := config.LoadProfiles()
		if err != nil {
			m.message = "Error: " + err.Error()
			m.isError = true
			m.state = viewMessage
			return m, nil
		}
		data, err := json.MarshalIndent(pf, "", "  ")
		if err != nil {
			m.message = "Error: " + err.Error()
			m.isError = true
			m.state = viewMessage
			return m, nil
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			m.formErr = err.Error()
			return m, textinput.Blink
		}
		m.message = fmt.Sprintf("Exported %d profile(s) to %s", len(pf.Profiles), path)
		m.isError = false
		m.state = viewMessage
		return m, nil
	default:
		m.formErr = ""
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateImportPath(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = viewList
		return m, nil
	case "enter":
		path := strings.TrimSpace(m.input.Value())
		if path == "" {
			m.formErr = "enter a file path"
			return m, textinput.Blink
		}
		data, err := os.ReadFile(path) // #nosec G304 -- path is user-provided via TUI input
		if err != nil {
			m.formErr = err.Error()
			return m, textinput.Blink
		}
		var pf config.ProfilesFile
		if err := json.Unmarshal(data, &pf); err != nil {
			m.formErr = "invalid file: " + err.Error()
			return m, textinput.Blink
		}
		if pf.Profiles == nil || len(pf.Profiles) == 0 {
			m.formErr = "no profiles found in file"
			return m, textinput.Blink
		}
		// Validate before importing
		for name, vars := range pf.Profiles {
			if err := validate.ProfileName(name); err != nil {
				m.formErr = fmt.Sprintf("invalid profile name %q: %v", name, err)
				return m, textinput.Blink
			}
			for k := range vars {
				if err := validate.EnvKey(k); err != nil {
					m.formErr = fmt.Sprintf("invalid key %q in profile %q: %v", k, name, err)
					return m, textinput.Blink
				}
			}
		}
		if m.importReplace {
			existing, err := config.LoadProfiles()
			if err != nil {
				m.message = "Error: " + err.Error()
				m.isError = true
				m.state = viewMessage
				return m, nil
			}
			for name := range existing.Profiles {
				if err := profile.DeleteProfile(name, true); err != nil {
					m.message = "Error deleting " + name + ": " + err.Error()
					m.isError = true
					m.state = viewMessage
					return m, nil
				}
			}
		}
		added := 0
		updated := 0
		for name, vars := range pf.Profiles {
			_, err := profile.GetProfile(name)
			exists := err == nil
			if err := profile.UpsertProfile(name, vars); err != nil {
				m.message = "Error importing " + name + ": " + err.Error()
				m.isError = true
				m.state = viewMessage
				return m, nil
			}
			if exists {
				updated++
			} else {
				added++
			}
		}
		m.reload()
		m.message = fmt.Sprintf("Imported %d profile(s) (%d added, %d updated)", len(pf.Profiles), added, updated)
		m.isError = false
		m.state = viewMessage
		return m, nil
	case "r", "R":
		m.importReplace = !m.importReplace
		return m, nil
	default:
		m.formErr = ""
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// ── View renderers ───────────────────────────────────────────────────────────

func (m model) View() string {
	switch m.state {
	case viewList:
		return m.renderList()
	case viewAddName:
		return m.renderAddName()
	case viewVarList:
		return m.renderVarList()
	case viewEditOneVar:
		return m.renderEditOneVar()
	case viewDeleteConfirm:
		return m.renderDeleteConfirm()
	case viewExportPath:
		return m.renderExportPath()
	case viewImportPath:
		return m.renderImportPath()
	case viewMessage:
		return m.renderMessage()
	}
	return ""
}

func (m model) renderList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("⚡ Switchy") + "\n")

	status := "Shell: " + m.shell
	if m.current != "" {
		status += "  •  Current: " + m.current
	}
	b.WriteString(dimStyle.Render(status) + "\n\n")

	if len(m.profiles) == 0 {
		b.WriteString(dimStyle.Render("No profiles yet.") + "\n")
		b.WriteString(dimStyle.Render("Press 'a' to create one.") + "\n")
	} else {
		for i, name := range m.profiles {
			isCurrent := name == m.current
			isSelected := i == m.cursor

			var marker string
			if isCurrent {
				marker = currentStyle.Render("*")
			} else {
				marker = " "
			}

			var line string
			if isSelected {
				line = selectedStyle.Render("> ") + " " + marker + " " + selectedStyle.Render(name)
			} else {
				line = "   " + marker + " " + name
			}
			b.WriteString(line + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓/jk navigate  a add  p dup  e edit  d delete  s use  x view  E export  I import  q quit"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderAddName() string {
	var b strings.Builder
	if m.duplicateSource != "" {
		b.WriteString(titleStyle.Render("Duplicate: "+m.duplicateSource) + "\n\n")
	} else {
		b.WriteString(titleStyle.Render("New Profile") + "\n\n")
	}
	b.WriteString(labelStyle.Render("Profile name:") + "\n")
	b.WriteString(m.input.View() + "\n")
	if m.formErr != "" {
		b.WriteString(errorStyle.Render("  "+m.formErr) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter confirm  esc cancel"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderVarList() string {
	var b strings.Builder

	if m.isAdding {
		b.WriteString(titleStyle.Render("New Profile: "+m.editTarget) + "\n\n")
	} else {
		b.WriteString(titleStyle.Render("Edit: "+m.editTarget) + "\n\n")
	}

	if len(m.varKeys) == 0 {
		b.WriteString(dimStyle.Render("No variables yet. Press 'a' to add one.") + "\n")
	} else {
		b.WriteString(labelStyle.Render("Variables:") + "\n")
		for i, k := range m.varKeys {
			v := m.pendingVars[k]
			if validate.IsSensitiveKey(k) {
				v = validate.MaskValue(v)
			}
			entry := fmt.Sprintf("%-28s = %s", k, v)
			if i == m.varCursor {
				b.WriteString(selectedStyle.Render("> "+entry) + "\n")
			} else {
				b.WriteString("  " + entry + "\n")
			}
		}
	}

	if m.formErr != "" {
		b.WriteString("\n" + errorStyle.Render("  "+m.formErr) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓/jk navigate  a add  enter edit  d delete  s save  esc cancel"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderEditOneVar() string {
	var b strings.Builder

	if m.isAdding {
		b.WriteString(titleStyle.Render("New Profile: "+m.editTarget) + "\n\n")
	} else {
		b.WriteString(titleStyle.Render("Edit: "+m.editTarget) + "\n\n")
	}

	if m.editingKey != "" {
		b.WriteString(labelStyle.Render("Edit variable:") + "\n")
	} else {
		b.WriteString(labelStyle.Render("Add variable:") + "\n")
	}
	b.WriteString(m.input.View() + "\n")

	if m.formErr != "" {
		b.WriteString(errorStyle.Render("  "+m.formErr) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter confirm  esc cancel"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderDeleteConfirm() string {
	var b strings.Builder
	b.WriteString(errorStyle.Render("Delete Profile") + "\n\n")
	fmt.Fprintf(&b, "Delete %q? This cannot be undone.\n\n", m.editTarget)
	b.WriteString(helpStyle.Render("y yes  n / esc cancel"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderExportPath() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Export Profiles") + "\n\n")
	b.WriteString(labelStyle.Render("File path:") + "\n")
	b.WriteString(m.input.View() + "\n")
	if m.formErr != "" {
		b.WriteString(errorStyle.Render("  "+m.formErr) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter confirm  esc cancel"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderImportPath() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Import Profiles") + "\n\n")
	b.WriteString(labelStyle.Render("File path:") + "\n")
	b.WriteString(m.input.View() + "\n")
	replaceLabel := "no"
	if m.importReplace {
		replaceLabel = "yes (will replace all profiles)"
	}
	b.WriteString(dimStyle.Render("Replace all: " + replaceLabel) + "\n")
	if m.formErr != "" {
		b.WriteString(errorStyle.Render("  "+m.formErr) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter confirm  r toggle replace  esc cancel"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderMessage() string {
	var b strings.Builder
	if m.isError {
		b.WriteString(errorStyle.Render("Error") + "\n\n")
		b.WriteString(m.message + "\n")
	} else {
		b.WriteString(successStyle.Render(m.message) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("press any key to continue"))
	b.WriteString("\n")
	return b.String()
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
