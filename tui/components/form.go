package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FieldType defines the type of form field.
type FieldType int

const (
	FieldText FieldType = iota
	FieldPassword
	FieldSelect
	FieldToggle
)

// Field represents a single form field.
type Field struct {
	Key         string
	Label       string
	Type        FieldType
	Value       string
	Options     []string // for FieldSelect
	Required    bool
	Placeholder string
	Disabled    bool
}

// FormModel is a generic form component.
type FormModel struct {
	fields      []Field
	inputs      []textinput.Model
	focused     int
	width       int
	height      int
	onSave      func(values map[string]string) tea.Cmd
	onCancel    func() tea.Cmd
	dirty       bool
	err         string

	// Styles
	labelStyle    lipgloss.Style
	inputStyle    lipgloss.Style
	focusedStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	disabledStyle lipgloss.Style
}

// NewForm creates a new form with the given fields.
func NewForm(fields []Field) FormModel {
	m := FormModel{
		fields: fields,
		inputs: make([]textinput.Model, len(fields)),
		labelStyle: lipgloss.NewStyle().
			Width(20).
			Foreground(lipgloss.Color("7")),
		inputStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")),
		focusedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Bold(true),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")),
		disabledStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
	}

	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		ti.Width = 40

		switch f.Type {
		case FieldPassword:
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '*'
		case FieldSelect, FieldToggle:
			// These don't use textinput, but we still create one for consistency
		}

		ti.SetValue(f.Value)
		if i == 0 && !f.Disabled {
			ti.Focus()
		}
		m.inputs[i] = ti
	}

	return m
}

// SetOnSave sets the save callback.
func (m *FormModel) SetOnSave(fn func(values map[string]string) tea.Cmd) {
	m.onSave = fn
}

// SetOnCancel sets the cancel callback.
func (m *FormModel) SetOnCancel(fn func() tea.Cmd) {
	m.onCancel = fn
}

// SetSize sets the form dimensions.
func (m *FormModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	inputWidth := width - 25
	if inputWidth < 20 {
		inputWidth = 20
	}
	for i := range m.inputs {
		m.inputs[i].Width = inputWidth
	}
}

// SetError sets an error message to display.
func (m *FormModel) SetError(err string) {
	m.err = err
}

// GetValues returns all field values as a map.
func (m FormModel) GetValues() map[string]string {
	values := make(map[string]string)
	for i, f := range m.fields {
		switch f.Type {
		case FieldSelect:
			values[f.Key] = m.fields[i].Value
		case FieldToggle:
			values[f.Key] = m.fields[i].Value
		default:
			values[f.Key] = m.inputs[i].Value()
		}
	}
	return values
}

// SetValue sets a field value by key.
func (m *FormModel) SetValue(key, value string) {
	for i, f := range m.fields {
		if f.Key == key {
			m.fields[i].Value = value
			if f.Type == FieldText || f.Type == FieldPassword {
				m.inputs[i].SetValue(value)
			}
			break
		}
	}
}

// IsDirty returns true if any field has been modified.
func (m FormModel) IsDirty() bool {
	return m.dirty
}

// Init implements tea.Model.
func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.focusNext()
			return m, nil
		case "shift+tab", "up":
			m.focusPrev()
			return m, nil
		case "ctrl+s":
			return m, m.save()
		case "esc":
			if m.onCancel != nil {
				return m, m.onCancel()
			}
			return m, nil
		case "enter":
			// For select fields, cycle through options
			if m.focused < len(m.fields) {
				f := &m.fields[m.focused]
				switch f.Type {
				case FieldSelect:
					m.cycleSelectOption(m.focused, 1)
					m.dirty = true
					return m, nil
				case FieldToggle:
					m.toggleField(m.focused)
					m.dirty = true
					return m, nil
				}
			}
		case "left":
			if m.focused < len(m.fields) && m.fields[m.focused].Type == FieldSelect {
				m.cycleSelectOption(m.focused, -1)
				m.dirty = true
				return m, nil
			}
		case "right":
			if m.focused < len(m.fields) && m.fields[m.focused].Type == FieldSelect {
				m.cycleSelectOption(m.focused, 1)
				m.dirty = true
				return m, nil
			}
		case " ":
			if m.focused < len(m.fields) && m.fields[m.focused].Type == FieldToggle {
				m.toggleField(m.focused)
				m.dirty = true
				return m, nil
			}
		}
	}

	// Update text inputs
	if m.focused < len(m.inputs) {
		f := m.fields[m.focused]
		if f.Type == FieldText || f.Type == FieldPassword {
			var cmd tea.Cmd
			oldVal := m.inputs[m.focused].Value()
			m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
			if m.inputs[m.focused].Value() != oldVal {
				m.dirty = true
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m *FormModel) focusNext() {
	m.inputs[m.focused].Blur()
	start := m.focused
	m.focused++
	for m.focused < len(m.fields) && m.fields[m.focused].Disabled {
		m.focused++
	}
	if m.focused >= len(m.fields) {
		m.focused = 0
		for m.focused < len(m.fields) && m.fields[m.focused].Disabled {
			m.focused++
		}
	}
	if m.focused >= len(m.fields) {
		// All fields disabled, reset to original
		m.focused = start
		return
	}
	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Focus()
	}
}

func (m *FormModel) focusPrev() {
	m.inputs[m.focused].Blur()
	start := m.focused
	m.focused--
	for m.focused >= 0 && m.fields[m.focused].Disabled {
		m.focused--
	}
	if m.focused < 0 {
		m.focused = len(m.fields) - 1
		for m.focused >= 0 && m.fields[m.focused].Disabled {
			m.focused--
		}
	}
	if m.focused < 0 {
		// All fields disabled, reset to original
		m.focused = start
		return
	}
	if m.focused >= 0 && m.focused < len(m.inputs) {
		m.inputs[m.focused].Focus()
	}
}

func (m *FormModel) cycleSelectOption(idx, delta int) {
	f := &m.fields[idx]
	if len(f.Options) == 0 {
		return
	}
	current := 0
	for i, opt := range f.Options {
		if opt == f.Value {
			current = i
			break
		}
	}
	current += delta
	if current < 0 {
		current = len(f.Options) - 1
	}
	if current >= len(f.Options) {
		current = 0
	}
	f.Value = f.Options[current]
}

func (m *FormModel) toggleField(idx int) {
	f := &m.fields[idx]
	if f.Value == "true" {
		f.Value = "false"
	} else {
		f.Value = "true"
	}
}

func (m FormModel) save() tea.Cmd {
	// Validate required fields
	for i, f := range m.fields {
		if f.Required {
			var val string
			if f.Type == FieldSelect || f.Type == FieldToggle {
				val = f.Value
			} else {
				val = m.inputs[i].Value()
			}
			if strings.TrimSpace(val) == "" {
				m.err = f.Label + " is required"
				return nil
			}
		}
	}

	if m.onSave != nil {
		return m.onSave(m.GetValues())
	}
	return nil
}

// View implements tea.Model.
func (m FormModel) View() string {
	var b strings.Builder

	for i, f := range m.fields {
		isFocused := i == m.focused
		label := f.Label
		if f.Required {
			label += " *"
		}
		label += ":"

		// Label
		if isFocused {
			b.WriteString(m.focusedStyle.Render(label))
		} else if f.Disabled {
			b.WriteString(m.disabledStyle.Render(label))
		} else {
			b.WriteString(m.labelStyle.Render(label))
		}
		b.WriteString(" ")

		// Value
		switch f.Type {
		case FieldSelect:
			val := "[" + f.Value + " ▼]"
			if isFocused {
				b.WriteString(m.focusedStyle.Render(val))
			} else {
				b.WriteString(m.inputStyle.Render(val))
			}
		case FieldToggle:
			val := "[ ]"
			if f.Value == "true" {
				val = "[x]"
			}
			if isFocused {
				b.WriteString(m.focusedStyle.Render(val))
			} else {
				b.WriteString(m.inputStyle.Render(val))
			}
		default:
			b.WriteString(m.inputs[i].View())
		}

		b.WriteString("\n")
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(m.errorStyle.Render("Error: " + m.err))
		b.WriteString("\n")
	}

	return b.String()
}
