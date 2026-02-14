package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents a single item in the list.
type ListItem struct {
	ID       string
	Label    string
	Sublabel string
	Icon     string // optional prefix icon/indicator
	Disabled bool
	Reason   string // reason if disabled
}

// ListSection represents a collapsible section in the list.
type ListSection struct {
	Name      string
	Collapsed bool
	Items     []ListItem
}

// ListModel is a generic scrollable list component.
type ListModel struct {
	sections    []ListSection
	flatItems   []flatItem // flattened view for navigation
	cursor      int
	width       int
	height      int
	onSelect    func(sectionIdx, itemIdx int, item ListItem) tea.Cmd
	showCursor  bool
	singleSection bool // if true, don't show section headers

	// Styles
	titleStyle    lipgloss.Style
	itemStyle     lipgloss.Style
	selectedStyle lipgloss.Style
	sublabelStyle lipgloss.Style
	disabledStyle lipgloss.Style
}

// flatItem maps cursor position to section/item indices
type flatItem struct {
	sectionIdx int
	itemIdx    int  // -1 means this is a section header
	isHeader   bool
}

// NewList creates a new list component.
func NewList(sections []ListSection) ListModel {
	m := ListModel{
		sections:   sections,
		showCursor: true,
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")),
		itemStyle: lipgloss.NewStyle().
			PaddingLeft(2),
		selectedStyle: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("14")).
			Bold(true),
		sublabelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		disabledStyle: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("8")),
	}
	m.rebuildFlatItems()
	return m
}

// NewSimpleList creates a list with a single unnamed section.
func NewSimpleList(items []ListItem) ListModel {
	m := NewList([]ListSection{{Items: items}})
	m.singleSection = true
	return m
}

// SetOnSelect sets the callback for item selection.
func (m *ListModel) SetOnSelect(fn func(sectionIdx, itemIdx int, item ListItem) tea.Cmd) {
	m.onSelect = fn
}

// SetSize sets the list dimensions.
func (m *ListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetShowCursor controls cursor visibility.
func (m *ListModel) SetShowCursor(show bool) {
	m.showCursor = show
}

// GetSelectedItem returns the currently selected item.
func (m ListModel) GetSelectedItem() (sectionIdx, itemIdx int, item ListItem, ok bool) {
	if m.cursor < 0 || m.cursor >= len(m.flatItems) {
		return 0, 0, ListItem{}, false
	}
	fi := m.flatItems[m.cursor]
	if fi.isHeader {
		return fi.sectionIdx, -1, ListItem{}, false
	}
	if fi.sectionIdx >= len(m.sections) || fi.itemIdx >= len(m.sections[fi.sectionIdx].Items) {
		return 0, 0, ListItem{}, false
	}
	return fi.sectionIdx, fi.itemIdx, m.sections[fi.sectionIdx].Items[fi.itemIdx], true
}

// SetSections replaces all sections.
func (m *ListModel) SetSections(sections []ListSection) {
	m.sections = sections
	m.rebuildFlatItems()
	if m.cursor >= len(m.flatItems) {
		m.cursor = len(m.flatItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// ToggleSection collapses/expands a section.
func (m *ListModel) ToggleSection(sectionIdx int) {
	if sectionIdx >= 0 && sectionIdx < len(m.sections) {
		m.sections[sectionIdx].Collapsed = !m.sections[sectionIdx].Collapsed
		m.rebuildFlatItems()
	}
}

func (m *ListModel) rebuildFlatItems() {
	m.flatItems = nil
	for si, sec := range m.sections {
		if !m.singleSection {
			m.flatItems = append(m.flatItems, flatItem{sectionIdx: si, itemIdx: -1, isHeader: true})
		}
		if !sec.Collapsed {
			for ii := range sec.Items {
				m.flatItems = append(m.flatItems, flatItem{sectionIdx: si, itemIdx: ii, isHeader: false})
			}
		}
	}
}

// Init implements tea.Model.
func (m ListModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "enter", " ":
			return m, m.selectCurrent()
		case "tab":
			// Toggle section if on header
			if m.cursor >= 0 && m.cursor < len(m.flatItems) {
				fi := m.flatItems[m.cursor]
				if fi.isHeader {
					m.ToggleSection(fi.sectionIdx)
				}
			}
		}
	}
	return m, nil
}

func (m *ListModel) moveCursor(delta int) {
	if len(m.flatItems) == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.flatItems) {
		m.cursor = len(m.flatItems) - 1
	}
}

func (m ListModel) selectCurrent() tea.Cmd {
	if m.cursor < 0 || m.cursor >= len(m.flatItems) {
		return nil
	}
	fi := m.flatItems[m.cursor]
	if fi.isHeader {
		// Toggle section on enter
		m.ToggleSection(fi.sectionIdx)
		return nil
	}
	if fi.sectionIdx >= len(m.sections) || fi.itemIdx >= len(m.sections[fi.sectionIdx].Items) {
		return nil
	}
	item := m.sections[fi.sectionIdx].Items[fi.itemIdx]
	if item.Disabled {
		return nil
	}
	if m.onSelect != nil {
		return m.onSelect(fi.sectionIdx, fi.itemIdx, item)
	}
	return nil
}

// View implements tea.Model.
func (m ListModel) View() string {
	var b strings.Builder

	for i, fi := range m.flatItems {
		isSelected := m.showCursor && i == m.cursor

		if fi.isHeader {
			sec := m.sections[fi.sectionIdx]
			arrow := "▼"
			if sec.Collapsed {
				arrow = "▶"
			}
			header := arrow + " " + sec.Name
			if len(sec.Items) > 0 {
				header += fmt.Sprintf(" (%d)", len(sec.Items))
			}
			if isSelected {
				b.WriteString(m.selectedStyle.Render(header))
			} else {
				b.WriteString(m.titleStyle.Render(header))
			}
		} else {
			item := m.sections[fi.sectionIdx].Items[fi.itemIdx]
			line := item.Label
			if item.Icon != "" {
				line = item.Icon + " " + line
			}

			var style lipgloss.Style
			if item.Disabled {
				style = m.disabledStyle
			} else if isSelected {
				style = m.selectedStyle
				line = "> " + line
			} else {
				style = m.itemStyle
				line = "  " + line
			}

			b.WriteString(style.Render(line))

			if item.Sublabel != "" && !item.Disabled {
				b.WriteString(" ")
				b.WriteString(m.sublabelStyle.Render(item.Sublabel))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}
