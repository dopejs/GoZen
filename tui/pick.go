package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/anthropics/opencc/internal/envfile"
)

type pickModel struct {
	choices  []string
	selected map[int]bool
	cursor   int
	done     bool
	cancelled bool
}

func newPickModel() pickModel {
	names := envfile.ConfigNames()
	return pickModel{
		choices:  names,
		selected: make(map[int]bool),
	}
}

func (m pickModel) Init() tea.Cmd {
	return nil
}

func (m pickModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pickModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Select providers for this session:"))
	b.WriteString("\n\n")

	for i, name := range m.choices {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}

		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}

		style := dimStyle
		if i == m.cursor {
			style = selectedStyle
		}

		b.WriteString(style.Render(cursor + check + " " + name))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  space:toggle  enter:confirm  q:cancel"))

	return b.String()
}

// Result returns the selected provider names in order.
func (m pickModel) Result() []string {
	var result []string
	for i, name := range m.choices {
		if m.selected[i] {
			result = append(result, name)
		}
	}
	return result
}
