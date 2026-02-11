package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/anthropics/opencc/internal/config"
	"github.com/anthropics/opencc/internal/envfile"
)

type listModel struct {
	configs  []*envfile.Config
	fbOrder  map[string]int
	cursor   int
	status   string
	deleting bool // confirm delete mode
}

func newListModel() listModel {
	return listModel{}
}

type configsLoadedMsg struct {
	configs []*envfile.Config
	fbOrder map[string]int
}

func (m listModel) init() tea.Cmd {
	return func() tea.Msg {
		configs, _ := envfile.ListConfigs()
		fbNames, _ := config.ReadFallbackOrder()
		fbOrder := make(map[string]int)
		for i, n := range fbNames {
			fbOrder[n] = i + 1
		}
		return configsLoadedMsg{configs: configs, fbOrder: fbOrder}
	}
}

func (m listModel) update(msg tea.Msg) (listModel, tea.Cmd) {
	switch msg := msg.(type) {
	case configsLoadedMsg:
		m.configs = msg.configs
		m.fbOrder = msg.fbOrder
		m.cursor = 0
		m.deleting = false
		return m, nil

	case statusMsg:
		m.status = msg.text
		return m, nil

	case tea.KeyMsg:
		if m.deleting {
			return m.handleDeleteConfirm(msg)
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m listModel) handleKey(msg tea.KeyMsg) (listModel, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.configs)-1 {
			m.cursor++
		}
	case "a":
		// Add new config
		return m, func() tea.Msg { return switchToEditorMsg{} }
	case "e", "enter":
		if len(m.configs) > 0 {
			name := m.configs[m.cursor].Name
			return m, func() tea.Msg { return switchToEditorMsg{configName: name} }
		}
	case "d":
		if len(m.configs) > 0 {
			m.deleting = true
		}
	case "f":
		return m, func() tea.Msg { return switchToFallbackMsg{} }
	}
	return m, nil
}

func (m listModel) handleDeleteConfirm(msg tea.KeyMsg) (listModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.cursor < len(m.configs) {
			cfg := m.configs[m.cursor]
			cfg.Delete()
			m.deleting = false
			return m, m.init()
		}
	case "n", "N", "esc":
		m.deleting = false
	}
	return m, nil
}

func (m listModel) view(width, height int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  opencc configurations"))
	b.WriteString("\n\n")

	if len(m.configs) == 0 {
		b.WriteString("  No configurations found.\n")
		b.WriteString("  Press 'a' to add a new configuration.\n")
	} else {
		for i, cfg := range m.configs {
			cursor := "  "
			style := dimStyle
			if i == m.cursor {
				cursor = "▸ "
				style = selectedStyle
			}

			baseURL := cfg.Get("ANTHROPIC_BASE_URL")
			model := cfg.Get("ANTHROPIC_MODEL")
			if model == "" {
				model = "-"
			}

			fbTag := ""
			if idx, ok := m.fbOrder[cfg.Name]; ok {
				fbTag = fmt.Sprintf(" [fb:%d]", idx)
			}

			line := fmt.Sprintf("%s%-12s model=%-20s  %s%s", cursor, cfg.Name, model, baseURL, fbTag)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.deleting && m.cursor < len(m.configs) {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Delete '%s'? (y/n)", m.configs[m.cursor].Name)))
	} else {
		b.WriteString(helpStyle.Render("  a:add  e/enter:edit  d:delete  f:fallback order  q:quit"))
	}

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(successStyle.Render("  " + m.status))
	}

	return b.String()
}
