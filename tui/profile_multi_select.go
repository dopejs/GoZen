package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/gozen/internal/config"
)

// profileMultiSelectModel is a checkbox-based profile multi-select TUI.
type profileMultiSelectModel struct {
	profiles  []string
	selected  map[string]bool
	cursor    int
	done      bool
	cancelled bool
}

func newProfileMultiSelectModel() profileMultiSelectModel {
	names := config.ListProfiles()
	return profileMultiSelectModel{
		profiles: names,
		selected: make(map[string]bool),
	}
}

func (m profileMultiSelectModel) Init() tea.Cmd {
	return nil
}

func (m profileMultiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "esc", "q":
			m.done = true
			// esc/q = skip, no selections
			m.selected = make(map[string]bool)
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}
		case " ":
			if m.cursor < len(m.profiles) {
				name := m.profiles[m.cursor]
				if m.selected[name] {
					delete(m.selected, name)
				} else {
					m.selected[name] = true
				}
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m profileMultiSelectModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Add to profiles"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Space toggle, Enter confirm, Esc skip"))
	b.WriteString("\n\n")

	for i, name := range m.profiles {
		cursor := "  "
		style := dimStyle
		if i == m.cursor {
			cursor = "▸ "
			style = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
		}

		var checkbox string
		if m.selected[name] {
			checkbox = lipgloss.NewStyle().
				Foreground(successColor).
				Render("[✓]")
		} else {
			checkbox = dimStyle.Render("[ ]")
		}

		b.WriteString(style.Render(cursor+checkbox+" "+name))
		if i < len(m.profiles)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Result returns the selected profile names.
func (m profileMultiSelectModel) Result() []string {
	var result []string
	for _, name := range m.profiles {
		if m.selected[name] {
			result = append(result, name)
		}
	}
	return result
}

// RunProfileMultiSelect runs a standalone profile multi-select TUI.
// Returns selected profile names. Esc/q returns nil, nil (skip).
func RunProfileMultiSelect() ([]string, error) {
	m := newProfileMultiSelectModel()
	if len(m.profiles) == 0 {
		return nil, nil
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, err
	}
	pm := result.(profileMultiSelectModel)
	if pm.cancelled {
		return nil, fmt.Errorf("cancelled")
	}
	return pm.Result(), nil
}
