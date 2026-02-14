package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
)

// MenuAction represents a menu item action.
type MenuAction int

const (
	MenuLaunch MenuAction = iota
	MenuConfigure
	MenuSettings
	MenuWebUI
	MenuQuit
)

type menuItem struct {
	label  string
	action MenuAction
}

// MenuModel is the main menu screen.
type MenuModel struct {
	items   []menuItem
	cursor  int
	width   int
	height  int
	profile string
	cli     string

	// Styles
	titleStyle    lipgloss.Style
	itemStyle     lipgloss.Style
	selectedStyle lipgloss.Style
	statusStyle   lipgloss.Style
	boxStyle      lipgloss.Style
}

// NewMenuModel creates a new main menu.
func NewMenuModel() MenuModel {
	return MenuModel{
		items: []menuItem{
			{label: "Launch", action: MenuLaunch},
			{label: "Configure", action: MenuConfigure},
			{label: "Settings", action: MenuSettings},
			{label: "Web UI", action: MenuWebUI},
			{label: "Quit", action: MenuQuit},
		},
		profile: config.GetDefaultProfile(),
		cli:     config.GetDefaultCLI(),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")).
			MarginBottom(1),
		itemStyle: lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(lipgloss.Color("7")),
		selectedStyle: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("14")).
			Bold(true),
		statusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginTop(1),
		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1, 3),
	}
}

// Init implements tea.Model.
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// MenuSelectedMsg is sent when a menu item is selected.
type MenuSelectedMsg struct {
	Action MenuAction
}

// Update implements tea.Model.
func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m, func() tea.Msg {
				return MenuSelectedMsg{Action: m.items[m.cursor].action}
			}
		case "q", "esc":
			return m, tea.Quit
		case "1":
			m.cursor = 0
			return m, func() tea.Msg { return MenuSelectedMsg{Action: MenuLaunch} }
		case "2":
			m.cursor = 1
			return m, func() tea.Msg { return MenuSelectedMsg{Action: MenuConfigure} }
		case "3":
			m.cursor = 2
			return m, func() tea.Msg { return MenuSelectedMsg{Action: MenuSettings} }
		case "4":
			m.cursor = 3
			return m, func() tea.Msg { return MenuSelectedMsg{Action: MenuWebUI} }
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m MenuModel) View() string {
	// Title
	title := m.titleStyle.Render("OpenCC")
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Environment Switcher")

	// Menu items
	var menuItems string
	for i, item := range m.items {
		var line string
		if i == m.cursor {
			line = m.selectedStyle.Render(fmt.Sprintf("> %s", item.label))
		} else {
			line = m.itemStyle.Render(fmt.Sprintf("  %s", item.label))
		}
		menuItems += line + "\n"
	}

	// Status line
	status := m.statusStyle.Render(fmt.Sprintf("Profile: %s  |  CLI: %s", m.profile, m.cli))

	// Combine into box
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		subtitle,
		"",
		menuItems,
	)
	box := m.boxStyle.Render(content)

	// Center on screen
	boxWidth := lipgloss.Width(box)
	boxHeight := lipgloss.Height(box)

	padLeft := (m.width - boxWidth) / 2
	padTop := (m.height - boxHeight - 2) / 2

	if padLeft < 0 {
		padLeft = 0
	}
	if padTop < 0 {
		padTop = 0
	}

	// Build final view
	var view string
	for i := 0; i < padTop; i++ {
		view += "\n"
	}

	lines := strings.Split(box, "\n")
	for _, line := range lines {
		for i := 0; i < padLeft; i++ {
			view += " "
		}
		view += line + "\n"
	}

	// Status at bottom
	for i := 0; i < padLeft; i++ {
		view += " "
	}
	view += status

	return view
}

// Refresh reloads config values.
func (m *MenuModel) Refresh() {
	m.profile = config.GetDefaultProfile()
	m.cli = config.GetDefaultCLI()
}
