package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/anthropics/opencc/internal/config"
	"github.com/anthropics/opencc/internal/envfile"
)

type fallbackModel struct {
	allConfigs []string
	order      []string // current fallback order
	cursor     int
	status     string
}

func newFallbackModel() fallbackModel {
	return fallbackModel{}
}

type fallbackLoadedMsg struct {
	allConfigs []string
	order      []string
}

func (m fallbackModel) init() tea.Cmd {
	return func() tea.Msg {
		names := envfile.ConfigNames()
		order, _ := config.ReadFallbackOrder()
		return fallbackLoadedMsg{allConfigs: names, order: order}
	}
}

func (m fallbackModel) update(msg tea.Msg) (fallbackModel, tea.Cmd) {
	switch msg := msg.(type) {
	case fallbackLoadedMsg:
		m.allConfigs = msg.allConfigs
		m.order = msg.order
		m.cursor = 0
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m fallbackModel) handleKey(msg tea.KeyMsg) (fallbackModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return m, func() tea.Msg { return switchToListMsg{} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.order)-1 {
			m.cursor++
		}
	case "K", "shift+up":
		// Move up in order
		if m.cursor > 0 {
			m.order[m.cursor], m.order[m.cursor-1] = m.order[m.cursor-1], m.order[m.cursor]
			m.cursor--
		}
	case "J", "shift+down":
		// Move down in order
		if m.cursor < len(m.order)-1 {
			m.order[m.cursor], m.order[m.cursor+1] = m.order[m.cursor+1], m.order[m.cursor]
			m.cursor++
		}
	case "a":
		// Add a provider not yet in the order
		for _, name := range m.allConfigs {
			if !contains(m.order, name) {
				m.order = append(m.order, name)
				break
			}
		}
	case "d", "delete":
		// Remove from order
		if m.cursor < len(m.order) {
			m.order = append(m.order[:m.cursor], m.order[m.cursor+1:]...)
			if m.cursor >= len(m.order) && m.cursor > 0 {
				m.cursor--
			}
		}
	case "s", "ctrl+s":
		// Save
		if err := config.WriteFallbackOrder(m.order); err != nil {
			m.status = "Error: " + err.Error()
		} else {
			m.status = "Saved!"
		}
	}
	return m, nil
}

func (m fallbackModel) view(width, height int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Fallback Order"))
	b.WriteString("\n\n")

	if len(m.order) == 0 {
		b.WriteString("  No providers in fallback chain.\n")
		b.WriteString("  Press 'a' to add a provider.\n")
	} else {
		for i, name := range m.order {
			cursor := "  "
			style := dimStyle
			if i == m.cursor {
				cursor = "▸ "
				style = selectedStyle
			}
			line := fmt.Sprintf("%s[%d] %s", cursor, i+1, name)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	// Show available providers not in order
	var available []string
	for _, name := range m.allConfigs {
		if !contains(m.order, name) {
			available = append(available, name)
		}
	}
	if len(available) > 0 {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("  Available: %s", strings.Join(available, ", "))))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.status != "" {
		b.WriteString(successStyle.Render("  " + m.status))
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render("  j/k:move  J/K:reorder  a:add  d:remove  s:save  esc:back"))

	return b.String()
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
