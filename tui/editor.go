package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/anthropics/opencc/internal/config"
	"github.com/anthropics/opencc/internal/envfile"
)

type editorField int

const (
	fieldName editorField = iota
	fieldBaseURL
	fieldAuthToken
	fieldModel
	fieldCount
)

type editorModel struct {
	fields   [fieldCount]textinput.Model
	focus    editorField
	editing  string // config name being edited, empty = new
	err      string
}

func newEditorModel(configName string) editorModel {
	var fields [fieldCount]textinput.Model

	for i := range fields {
		fields[i] = textinput.New()
		fields[i].CharLimit = 256
	}

	fields[fieldName].Placeholder = "config name (e.g. work)"
	fields[fieldName].Prompt = "  Name:       "
	fields[fieldBaseURL].Placeholder = "https://api.example.com"
	fields[fieldBaseURL].Prompt = "  Base URL:   "
	fields[fieldAuthToken].Placeholder = "sk-..."
	fields[fieldAuthToken].Prompt = "  Auth Token: "
	fields[fieldAuthToken].EchoMode = textinput.EchoPassword
	fields[fieldModel].Placeholder = "claude-sonnet-4-20250514"
	fields[fieldModel].Prompt = "  Model:      "

	m := editorModel{
		fields:  fields,
		editing: configName,
	}

	if configName != "" {
		// Load existing config
		cfg, err := envfile.LoadByName(configName)
		if err == nil {
			m.fields[fieldName].SetValue(cfg.Name)
			m.fields[fieldBaseURL].SetValue(cfg.Get("ANTHROPIC_BASE_URL"))
			m.fields[fieldAuthToken].SetValue(cfg.Get("ANTHROPIC_AUTH_TOKEN"))
			m.fields[fieldModel].SetValue(cfg.Get("ANTHROPIC_MODEL"))
		}
		// Disable name field when editing
		m.focus = fieldBaseURL
	} else {
		m.focus = fieldName
	}

	m.fields[m.focus].Focus()
	return m
}

func (m editorModel) init() tea.Cmd {
	return textinput.Blink
}

func (m editorModel) update(msg tea.Msg) (editorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchToListMsg{} }
		case "tab", "down":
			m.fields[m.focus].Blur()
			m.focus = (m.focus + 1) % fieldCount
			if m.editing != "" && m.focus == fieldName {
				m.focus = fieldBaseURL
			}
			m.fields[m.focus].Focus()
			return m, textinput.Blink
		case "shift+tab", "up":
			m.fields[m.focus].Blur()
			m.focus = (m.focus - 1 + fieldCount) % fieldCount
			if m.editing != "" && m.focus == fieldName {
				m.focus = fieldModel
			}
			m.fields[m.focus].Focus()
			return m, textinput.Blink
		case "ctrl+s", "enter":
			if m.focus == fieldCount-1 || msg.String() == "ctrl+s" {
				return m.save()
			}
			// Enter on non-last field = move to next
			m.fields[m.focus].Blur()
			m.focus = (m.focus + 1) % fieldCount
			if m.editing != "" && m.focus == fieldName {
				m.focus = fieldBaseURL
			}
			m.fields[m.focus].Focus()
			return m, textinput.Blink
		}
	}

	// Update focused field
	var cmd tea.Cmd
	m.fields[m.focus], cmd = m.fields[m.focus].Update(msg)
	return m, cmd
}

func (m editorModel) save() (editorModel, tea.Cmd) {
	name := strings.TrimSpace(m.fields[fieldName].Value())
	baseURL := strings.TrimSpace(m.fields[fieldBaseURL].Value())
	token := strings.TrimSpace(m.fields[fieldAuthToken].Value())
	model := strings.TrimSpace(m.fields[fieldModel].Value())

	if name == "" {
		m.err = "name is required"
		return m, nil
	}
	if baseURL == "" {
		m.err = "base URL is required"
		return m, nil
	}
	if token == "" {
		m.err = "auth token is required"
		return m, nil
	}

	entries := []envfile.Entry{
		{Key: "ANTHROPIC_BASE_URL", Value: baseURL},
		{Key: "ANTHROPIC_AUTH_TOKEN", Value: token},
	}
	if model != "" {
		entries = append(entries, envfile.Entry{Key: "ANTHROPIC_MODEL", Value: model})
	}

	if m.editing != "" {
		// Update existing
		cfg, err := envfile.LoadByName(m.editing)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		cfg.Entries = entries
		if err := cfg.Save(); err != nil {
			m.err = err.Error()
			return m, nil
		}
	} else {
		// Create new
		if _, err := envfile.Create(name, entries); err != nil {
			m.err = err.Error()
			return m, nil
		}
		// Append to fallback.conf
		fbOrder, _ := config.ReadFallbackOrder()
		fbOrder = append(fbOrder, name)
		config.WriteFallbackOrder(fbOrder)
	}

	return m, func() tea.Msg { return switchToListMsg{} }
}

func (m editorModel) view(width, height int) string {
	var b strings.Builder

	title := "Add Configuration"
	if m.editing != "" {
		title = fmt.Sprintf("Edit Configuration: %s", m.editing)
	}
	b.WriteString(titleStyle.Render("  " + title))
	b.WriteString("\n\n")

	for i := range m.fields {
		if m.editing != "" && editorField(i) == fieldName {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  Name:       %s", m.editing)))
			b.WriteString("\n")
			continue
		}
		b.WriteString(m.fields[i].View())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.err != "" {
		b.WriteString(errorStyle.Render("  " + m.err))
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render("  tab:next field  ctrl+s:save  esc:cancel"))

	return b.String()
}
