package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type view int

const (
	viewList view = iota
	viewEditor
	viewFallback
)

type model struct {
	currentView view
	list        listModel
	editor      editorModel
	fallback    fallbackModel
	width       int
	height      int
	err         error
}

func initialModel() model {
	return model{
		currentView: viewList,
		list:        newListModel(),
		fallback:    newFallbackModel(),
	}
}

func (m model) Init() tea.Cmd {
	return m.list.init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.currentView {
	case viewList:
		return m.updateList(msg)
	case viewEditor:
		return m.updateEditor(msg)
	case viewFallback:
		return m.updateFallback(msg)
	}
	return m, nil
}

func (m model) View() string {
	switch m.currentView {
	case viewList:
		return m.list.view(m.width, m.height)
	case viewEditor:
		return m.editor.view(m.width, m.height)
	case viewFallback:
		return m.fallback.view(m.width, m.height)
	}
	return ""
}

// Run starts the TUI application.
func Run() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Styles
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
)

// Messages
type switchToListMsg struct{}
type switchToEditorMsg struct {
	configName string // empty = new config
}
type switchToFallbackMsg struct{}
type statusMsg struct {
	text string
}

func (m model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case switchToEditorMsg:
		m.currentView = viewEditor
		m.editor = newEditorModel(msg.configName)
		return m, m.editor.init()
	case switchToFallbackMsg:
		m.currentView = viewFallback
		m.fallback = newFallbackModel()
		return m, m.fallback.init()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.update(msg)
	return m, cmd
}

func (m model) updateEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case switchToListMsg:
		m.currentView = viewList
		m.list = newListModel()
		return m, m.list.init()
	}

	var cmd tea.Cmd
	m.editor, cmd = m.editor.update(msg)
	return m, cmd
}

func (m model) updateFallback(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case switchToListMsg:
		m.currentView = viewList
		m.list = newListModel()
		return m, m.list.init()
	}

	var cmd tea.Cmd
	m.fallback, cmd = m.fallback.update(msg)
	return m, cmd
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Placeholder to avoid unused import
var _ = fmt.Sprintf
