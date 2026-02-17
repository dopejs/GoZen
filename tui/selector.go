package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// selectorItem represents an item in the selector list.
type selectorItem struct {
	name     string
	hint     string // shown grayed out after name, e.g. "(current)"
	disabled bool
	reason   string // reason for being disabled
}

// selectorModel is a simple list selector TUI.
type selectorModel struct {
	title     string
	items     []selectorItem
	cursor    int
	selected  string
	cancelled bool
}

func newSelectorModel(title string, items []selectorItem) selectorModel {
	m := selectorModel{
		title: title,
		items: items,
	}
	// Move cursor to first enabled item
	for i, item := range items {
		if !item.disabled {
			m.cursor = i
			break
		}
	}
	return m
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "enter":
			if m.cursor < len(m.items) && !m.items[m.cursor].disabled {
				m.selected = m.items[m.cursor].name
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *selectorModel) moveCursor(delta int) {
	if len(m.items) == 0 {
		return
	}

	// Find next enabled item in direction
	newCursor := m.cursor
	for {
		newCursor += delta
		if newCursor < 0 || newCursor >= len(m.items) {
			return // hit boundary, don't move
		}
		if !m.items[newCursor].disabled {
			m.cursor = newCursor
			return
		}
	}
}

func (m selectorModel) View() string {
	width := 80  // default width
	height := 24 // default height

	sidePadding := 2
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("☰ " + m.title)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Items box
	var content strings.Builder
	for i, item := range m.items {
		cursor := "  "
		style := tableRowStyle
		if i == m.cursor && !item.disabled {
			cursor = "▸ "
			style = tableSelectedRowStyle
		}
		if item.disabled {
			style = dimStyle
		}

		line := fmt.Sprintf("%s%s", cursor, item.name)
		if item.hint != "" {
			line += dimStyle.Render(" " + item.hint)
		}
		if item.disabled && item.reason != "" {
			line += dimStyle.Render(fmt.Sprintf(" (%s)", item.reason))
		}
		content.WriteString(style.Render(line))
		if i < len(m.items)-1 {
			content.WriteString("\n")
		}
	}

	itemBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(content.String())
	b.WriteString(itemBox)

	// Build view with side padding
	mainContent := b.String()
	var view strings.Builder
	lines := strings.Split(mainContent, "\n")
	for _, line := range lines {
		view.WriteString(strings.Repeat(" ", sidePadding))
		view.WriteString(line)
		view.WriteString("\n")
	}

	// Fill remaining space to push help bar to bottom
	currentLines := len(lines)
	remainingLines := height - currentLines - 1
	for i := 0; i < remainingLines; i++ {
		view.WriteString("\n")
	}

	// Help bar at bottom
	helpBar := RenderHelpBar("↑↓ move • Enter select • Esc cancel", width)
	view.WriteString(helpBar)

	return view.String()
}

// RunSelector runs a selector TUI and returns the selected item name.
func RunSelector(title string, items []selectorItem) (string, error) {
	m := newSelectorModel(title, items)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	sm := result.(selectorModel)
	if sm.cancelled {
		return "", fmt.Errorf("cancelled")
	}
	return sm.selected, nil
}

// minimalSelectorModel is a borderless radio-button selector.
// It uses alt screen so the list vanishes on exit.
type minimalSelectorModel struct {
	items     []string
	current   string // item marked with ● (the active value)
	cursor    int
	selected  string
	cancelled bool
}

func (m minimalSelectorModel) Init() tea.Cmd { return nil }

func (m minimalSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.items[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m minimalSelectorModel) View() string {
	var b strings.Builder
	for i, item := range m.items {
		dot := "○"
		if item == m.current {
			dot = "●"
		}
		line := fmt.Sprintf("%s %s", dot, item)
		if i == m.cursor {
			b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(line))
		} else {
			b.WriteString(dimStyle.Render(line))
		}
		if i < len(m.items)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// RunMinimalSelector runs a borderless radio-button selector in alt screen.
// current marks the active item with ●. Returns the selected item name,
// or "cancelled" error on esc/ctrl-c.
func RunMinimalSelector(items []string, current string) (string, error) {
	cursor := 0
	for i, item := range items {
		if item == current {
			cursor = i
			break
		}
	}
	m := minimalSelectorModel{
		items:   items,
		current: current,
		cursor:  cursor,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	sm := result.(minimalSelectorModel)
	if sm.cancelled {
		return "", fmt.Errorf("cancelled")
	}
	return sm.selected, nil
}
