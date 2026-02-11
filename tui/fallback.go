package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/anthropics/opencc/internal/config"
	"github.com/anthropics/opencc/internal/envfile"
)

type fallbackModel struct {
	profile    string   // profile name ("default", "work", etc.)
	allConfigs []string
	order      []string // current fallback order
	cursor     int
	grabbed    bool // true = item is grabbed and arrow keys reorder
	status     string
}

func newFallbackModel(profile string) fallbackModel {
	if profile == "" {
		profile = "default"
	}
	return fallbackModel{profile: profile}
}

type fallbackLoadedMsg struct {
	allConfigs []string
	order      []string
}

func (m fallbackModel) init() tea.Cmd {
	profile := m.profile
	return func() tea.Msg {
		names := envfile.ConfigNames()
		order, _ := config.ReadProfileOrder(profile)
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
	if m.grabbed {
		return m.handleGrabbed(msg)
	}

	switch msg.String() {
	case "esc", "q":
		config.WriteProfileOrder(m.profile, m.validOrder())
		return m, func() tea.Msg { return switchToListMsg{} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.order)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.order) > 0 {
			m.grabbed = true
		}
	case "a":
		for _, name := range m.allConfigs {
			if !contains(m.order, name) {
				m.order = append(m.order, name)
				break
			}
		}
	case "d", "delete":
		if m.cursor < len(m.order) {
			m.order = append(m.order[:m.cursor], m.order[m.cursor+1:]...)
			if m.cursor >= len(m.order) && m.cursor > 0 {
				m.cursor--
			}
		}
	case "s", "ctrl+s":
		m.order = m.validOrder()
		if err := config.WriteProfileOrder(m.profile, m.order); err != nil {
			m.status = "Error: " + err.Error()
		} else {
			m.status = "Saved!"
		}
		if m.cursor >= len(m.order) && m.cursor > 0 {
			m.cursor = len(m.order) - 1
		}
	}
	return m, nil
}

func (m fallbackModel) handleGrabbed(msg tea.KeyMsg) (fallbackModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.grabbed = false
	case "up", "k":
		if m.cursor > 0 {
			m.order[m.cursor], m.order[m.cursor-1] = m.order[m.cursor-1], m.order[m.cursor]
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.order)-1 {
			m.order[m.cursor], m.order[m.cursor+1] = m.order[m.cursor+1], m.order[m.cursor]
			m.cursor++
		}
	}
	return m, nil
}

// validOrder returns order with missing providers removed.
func (m fallbackModel) validOrder() []string {
	var valid []string
	for _, name := range m.order {
		if contains(m.allConfigs, name) {
			valid = append(valid, name)
		}
	}
	return valid
}

// strikethrough applies Unicode combining strikethrough to each rune.
func strikethrough(s string) string {
	var b strings.Builder
	for _, r := range s {
		b.WriteRune(r)
		b.WriteRune('\u0336')
	}
	return b.String()
}

func (m fallbackModel) view(width, height int) string {
	var b strings.Builder

	title := "Fallback Order"
	if m.profile != "" && m.profile != "default" {
		title = fmt.Sprintf("Fallback Order [%s]", m.profile)
	}
	b.WriteString(titleStyle.Render("  " + title))
	b.WriteString("\n\n")

	if len(m.order) == 0 {
		b.WriteString("  No providers in fallback chain.\n")
		b.WriteString("  Press 'a' to add a provider.\n")
	} else {
		for i, name := range m.order {
			cursor := "  "
			style := dimStyle
			missing := !contains(m.allConfigs, name)
			if i == m.cursor {
				if m.grabbed {
					cursor = "⇕ "
					style = grabbedStyle
				} else {
					cursor = "▸ "
					style = selectedStyle
				}
			}
			label := name
			if missing {
				label = strikethrough(name)
				if i != m.cursor {
					style = errorStyle
				}
			}
			line := fmt.Sprintf("%s[%d] %s", cursor, i+1, label)
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
	if m.grabbed {
		b.WriteString(helpStyle.Render("  ↑/↓:reorder  enter/esc:drop"))
	} else {
		b.WriteString(helpStyle.Render("  ↑/↓:move  enter:grab  a:add  d:remove  s:save  esc:back"))
	}

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
