package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/gozen/internal/config"
)

type pickModel struct {
	choices   []string // all available providers
	order     []string // selected providers in order
	cursor    int
	grabbed   bool
	done      bool
	cancelled bool
}

func newPickModel() pickModel {
	names := config.ProviderNames()
	return pickModel{
		choices: names,
	}
}

func (m pickModel) Init() tea.Cmd {
	return nil
}

// orderIndex returns the 1-based index in order, or 0 if not selected.
func (m pickModel) orderIndex(name string) int {
	for i, n := range m.order {
		if n == name {
			return i + 1
		}
	}
	return 0
}

func (m pickModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.grabbed {
		return m.updateGrabbed(msg)
	}

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
			// choices + 1 for the submit button
			if m.cursor < len(m.choices) {
				m.cursor++
			}
		case " ":
			if m.cursor < len(m.choices) {
				name := m.choices[m.cursor]
				if idx := m.orderIndex(name); idx > 0 {
					m.order = removeFromOrder(m.order, name)
				} else {
					m.order = append(m.order, name)
				}
			}
		case "enter":
			// On submit button or no selected item grabbed
			if m.cursor == len(m.choices) {
				m.done = true
				return m, tea.Quit
			}
			if m.cursor < len(m.choices) {
				name := m.choices[m.cursor]
				if m.orderIndex(name) > 0 {
					m.grabbed = true
					return m, nil
				}
			}
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pickModel) updateGrabbed(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.choices) {
		m.grabbed = false
		return m, nil
	}
	name := m.choices[m.cursor]
	orderIdx := m.orderIndex(name)
	if orderIdx == 0 {
		m.grabbed = false
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			m.grabbed = false
		case "up", "k":
			if orderIdx > 1 {
				m.order[orderIdx-1], m.order[orderIdx-2] = m.order[orderIdx-2], m.order[orderIdx-1]
			}
		case "down", "j":
			if orderIdx < len(m.order) {
				m.order[orderIdx-1], m.order[orderIdx] = m.order[orderIdx], m.order[orderIdx-1]
			}
		}
	}
	return m, nil
}

func (m pickModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Providers"))
	b.WriteString("\n")
	if m.grabbed {
		b.WriteString(dimStyle.Render("  ↑↓ reorder, Enter/Esc drop"))
	} else {
		b.WriteString(dimStyle.Render("  Space toggle, Enter reorder/confirm, Esc cancel"))
	}
	b.WriteString("\n\n")

	for i, name := range m.choices {
		cursor := "  "
		style := dimStyle
		orderIdx := m.orderIndex(name)

		if i == m.cursor {
			if m.grabbed {
				cursor = "⇕ "
				style = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
			} else {
				cursor = "▸ "
				style = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
			}
		}

		var checkbox string
		if orderIdx > 0 {
			checkbox = lipgloss.NewStyle().
				Foreground(successColor).
				Render(fmt.Sprintf("[%d]", orderIdx))
		} else {
			checkbox = dimStyle.Render("[ ]")
		}

		grabIndicator := ""
		if m.grabbed && i == m.cursor {
			grabIndicator = " " + lipgloss.NewStyle().
				Foreground(accentColor).
				Render("(reordering)")
		}

		line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, name, grabIndicator)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Submit button
	b.WriteString("\n")
	if m.cursor == len(m.choices) {
		btn := lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("▸ [ Start Session ]")
		b.WriteString(btn)
	} else if len(m.order) > 0 {
		btn := dimStyle.Render("  [ Start Session ]")
		b.WriteString(btn)
	} else {
		btn := dimStyle.Render("  [ Select providers to start ]")
		b.WriteString(btn)
	}

	return b.String()
}

// Result returns the selected provider names in order.
func (m pickModel) Result() []string {
	return m.order
}
