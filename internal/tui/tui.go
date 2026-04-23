package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(purple)
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
	viewList viewState = iota
	viewAddName
	viewEditVars
	viewDeleteConfirm
	viewMessage
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

	message string
	isError bool

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
	_, err := p.Run()
	return err
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
		case viewEditVars:
			return m.updateEditVars(msg)
		case viewDeleteConfirm:
			return m.updateDeleteConfirm(msg)
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
		for k, v := range vars {
			m.pendingVars[k] = v
		}
		m.state = viewEditVars
		m.input.Reset()
		m.input.Placeholder = "KEY=VALUE"
		m.input.Focus()
		m.formErr = ""
		return m, textinput.Blink

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
		if err := config.SaveState(&config.StateFile{
			CurrentProfile:     name,
			LastActivationMode: "session",
		}); err != nil {
			m.message = "Error: " + err.Error()
			m.isError = true
		} else {
			m.current = name
			m.message = fmt.Sprintf(
				"Profile %q selected.\n\nTo apply in your current shell:\n  eval \"$(swy export %s)\"",
				name, name,
			)
			m.isError = false
		}
		m.state = viewMessage

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
	}
	return m, nil
}

func (m model) updateAddName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = viewList
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.input.Value())
		if err := validate.ProfileName(name); err != nil {
			m.formErr = err.Error()
			return m, textinput.Blink
		}
		m.editTarget = name
		m.state = viewEditVars
		m.input.Reset()
		m.input.Placeholder = "KEY=VALUE"
		m.formErr = ""
		return m, textinput.Blink
	default:
		m.formErr = ""
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateEditVars(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = viewList
		m.reload()
		return m, nil
	case "enter":
		val := strings.TrimSpace(m.input.Value())
		if val == "" {
			if m.isAdding && len(m.pendingVars) == 0 {
				m.formErr = "at least one KEY=VALUE is required"
				return m, textinput.Blink
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
		pairs, err := validate.ParseKVPairs([]string{val})
		if err != nil {
			m.formErr = err.Error()
			m.input.Reset()
			return m, textinput.Blink
		}
		for k, v := range pairs {
			m.pendingVars[k] = v
		}
		m.formErr = ""
		m.input.Reset()
		return m, textinput.Blink
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

// ── View renderers ───────────────────────────────────────────────────────────

func (m model) View() string {
	switch m.state {
	case viewList:
		return m.renderList()
	case viewAddName:
		return m.renderAddName()
	case viewEditVars:
		return m.renderEditVars()
	case viewDeleteConfirm:
		return m.renderDeleteConfirm()
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
	b.WriteString(helpStyle.Render("↑↓/jk navigate  a add  e edit  d delete  s switch  x export  q quit"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderAddName() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("New Profile") + "\n\n")
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

func (m model) renderEditVars() string {
	var b strings.Builder

	if m.isAdding {
		b.WriteString(titleStyle.Render("New Profile: "+m.editTarget) + "\n\n")
	} else {
		b.WriteString(titleStyle.Render("Edit: "+m.editTarget) + "\n\n")
	}

	if len(m.pendingVars) > 0 {
		if m.isAdding {
			b.WriteString(labelStyle.Render("Variables:") + "\n")
		} else {
			b.WriteString(labelStyle.Render("Current variables:") + "\n")
		}
		for _, k := range sortedKeys(m.pendingVars) {
			v := m.pendingVars[k]
			if validate.IsSensitiveKey(k) {
				v = validate.MaskValue(v)
			}
			b.WriteString(fmt.Sprintf("  %-28s = %s\n", k, v))
		}
		b.WriteString("\n")
	}

	prompt := "Add/update KEY=VALUE (empty to save):"
	if m.isAdding {
		prompt = "Enter KEY=VALUE (empty to finish):"
	}
	b.WriteString(dimStyle.Render(prompt) + "\n")
	b.WriteString(m.input.View() + "\n")

	if m.formErr != "" {
		b.WriteString(errorStyle.Render("  "+m.formErr) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter add variable  empty enter saves  esc cancel"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderDeleteConfirm() string {
	var b strings.Builder
	b.WriteString(errorStyle.Render("Delete Profile") + "\n\n")
	b.WriteString(fmt.Sprintf("Delete %q? This cannot be undone.\n\n", m.editTarget))
	b.WriteString(helpStyle.Render("y yes  n / esc cancel"))
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
