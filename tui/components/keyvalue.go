package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KeyValuePair represents a single key-value entry.
type KeyValuePair struct {
	Key   string
	Value string
}

// KeyValueModel is a key-value editor component.
type KeyValueModel struct {
	pairs      []KeyValuePair
	cursor     int
	editing    bool
	editingKey bool // true = editing key, false = editing value
	keyInput   textinput.Model
	valueInput textinput.Model
	adding     bool
	width      int
	height     int
	onSave     func(pairs []KeyValuePair) tea.Cmd

	// Styles
	keyStyle      lipgloss.Style
	valueStyle    lipgloss.Style
	selectedStyle lipgloss.Style
	headerStyle   lipgloss.Style
}

// NewKeyValue creates a new key-value editor.
func NewKeyValue(pairs []KeyValuePair) KeyValueModel {
	keyInput := textinput.New()
	keyInput.Placeholder = "KEY"
	keyInput.Width = 30

	valueInput := textinput.New()
	valueInput.Placeholder = "value"
	valueInput.Width = 40

	return KeyValueModel{
		pairs:      pairs,
		keyInput:   keyInput,
		valueInput: valueInput,
		keyStyle: lipgloss.NewStyle().
			Width(25).
			Foreground(lipgloss.Color("12")),
		valueStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")),
		selectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Bold(true),
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("8")),
	}
}

// SetOnSave sets the save callback.
func (m *KeyValueModel) SetOnSave(fn func(pairs []KeyValuePair) tea.Cmd) {
	m.onSave = fn
}

// SetSize sets the component dimensions.
func (m *KeyValueModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetPairs returns all key-value pairs.
func (m KeyValueModel) GetPairs() []KeyValuePair {
	return m.pairs
}

// SetPairs replaces all pairs.
func (m *KeyValueModel) SetPairs(pairs []KeyValuePair) {
	m.pairs = pairs
	if m.cursor >= len(m.pairs) {
		m.cursor = len(m.pairs) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// Init implements tea.Model.
func (m KeyValueModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m KeyValueModel) Update(msg tea.Msg) (KeyValueModel, tea.Cmd) {
	if m.editing || m.adding {
		return m.updateEditing(msg)
	}
	return m.updateNormal(msg)
}

func (m KeyValueModel) updateNormal(msg tea.Msg) (KeyValueModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.pairs)-1 {
				m.cursor++
			}
		case "a":
			m.adding = true
			m.editingKey = true
			m.keyInput.SetValue("")
			m.valueInput.SetValue("")
			m.keyInput.Focus()
			return m, textinput.Blink
		case "e", "enter":
			if len(m.pairs) > 0 {
				m.editing = true
				m.editingKey = true
				m.keyInput.SetValue(m.pairs[m.cursor].Key)
				m.valueInput.SetValue(m.pairs[m.cursor].Value)
				m.keyInput.Focus()
				return m, textinput.Blink
			}
		case "d":
			if len(m.pairs) > 0 {
				m.pairs = append(m.pairs[:m.cursor], m.pairs[m.cursor+1:]...)
				if m.cursor >= len(m.pairs) && m.cursor > 0 {
					m.cursor--
				}
			}
		case "ctrl+s":
			if m.onSave != nil {
				return m, m.onSave(m.pairs)
			}
		}
	}
	return m, nil
}

func (m KeyValueModel) updateEditing(msg tea.Msg) (KeyValueModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.editing = false
			m.adding = false
			m.keyInput.Blur()
			m.valueInput.Blur()
			return m, nil
		case "tab":
			if m.editingKey {
				m.editingKey = false
				m.keyInput.Blur()
				m.valueInput.Focus()
			} else {
				m.editingKey = true
				m.valueInput.Blur()
				m.keyInput.Focus()
			}
			return m, nil
		case "enter":
			key := strings.TrimSpace(m.keyInput.Value())
			value := m.valueInput.Value()
			if key != "" {
				if m.adding {
					m.pairs = append(m.pairs, KeyValuePair{Key: key, Value: value})
					m.cursor = len(m.pairs) - 1
				} else {
					m.pairs[m.cursor] = KeyValuePair{Key: key, Value: value}
				}
			}
			m.editing = false
			m.adding = false
			m.keyInput.Blur()
			m.valueInput.Blur()
			return m, nil
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.editingKey {
		m.keyInput, cmd = m.keyInput.Update(msg)
	} else {
		m.valueInput, cmd = m.valueInput.Update(msg)
	}
	return m, cmd
}

// View implements tea.Model.
func (m KeyValueModel) View() string {
	var b strings.Builder

	if m.editing || m.adding {
		title := "Edit Entry"
		if m.adding {
			title = "Add Entry"
		}
		b.WriteString(m.headerStyle.Render(title))
		b.WriteString("\n\n")
		b.WriteString("Key:   ")
		b.WriteString(m.keyInput.View())
		b.WriteString("\n")
		b.WriteString("Value: ")
		b.WriteString(m.valueInput.View())
		b.WriteString("\n\n")
		b.WriteString(m.headerStyle.Render("[Enter] save  [Tab] switch  [Esc] cancel"))
		return b.String()
	}

	if len(m.pairs) == 0 {
		b.WriteString(m.headerStyle.Render("No entries. Press [a] to add."))
		return b.String()
	}

	for i, p := range m.pairs {
		isSelected := i == m.cursor
		key := p.Key
		value := p.Value

		if isSelected {
			b.WriteString(m.selectedStyle.Render("> " + key))
			b.WriteString(" = ")
			b.WriteString(m.selectedStyle.Render(value))
		} else {
			b.WriteString("  ")
			b.WriteString(m.keyStyle.Render(key))
			b.WriteString(" = ")
			b.WriteString(m.valueStyle.Render(value))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.headerStyle.Render("[a]dd [e]dit [d]elete [Ctrl+S] save"))

	return b.String()
}
