package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/gozen/internal/config"
)

// LaunchBackMsg is sent when user wants to go back from launch wizard.
type LaunchBackMsg struct{}

// LaunchStartMsg is sent when user confirms launch.
type LaunchStartMsg struct {
	Profile string
	CLI     string
}

// LaunchModel is the launch wizard screen.
type LaunchModel struct {
	profiles        []string
	clis            []string
	profileCursor   int
	cliCursor       int
	focusOnCLI      bool
	selectedProfile string
	selectedCLI     string
	width           int
	height          int
}

// NewLaunchModel creates a new launch wizard.
func NewLaunchModel() LaunchModel {
	profiles := config.ListProfiles()
	defaultProfile := config.GetDefaultProfile()
	defaultCLI := config.GetDefaultCLI()

	// Find default profile index
	profileIdx := 0
	for i, p := range profiles {
		if p == defaultProfile {
			profileIdx = i
			break
		}
	}

	// Find default CLI index
	clis := config.AvailableCLIs
	cliIdx := 0
	for i, c := range clis {
		if c == defaultCLI {
			cliIdx = i
			break
		}
	}

	return LaunchModel{
		profiles:        profiles,
		clis:            clis,
		profileCursor:   profileIdx,
		cliCursor:       cliIdx,
		selectedProfile: defaultProfile,
		selectedCLI:     defaultCLI,
	}
}

// Init implements tea.Model.
func (m LaunchModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m LaunchModel) Update(msg tea.Msg) (LaunchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.focusOnCLI {
				if m.cliCursor > 0 {
					m.cliCursor--
				}
			} else {
				if m.profileCursor > 0 {
					m.profileCursor--
				}
			}
		case "down", "j":
			if m.focusOnCLI {
				if m.cliCursor < len(m.clis)-1 {
					m.cliCursor++
				}
			} else {
				if m.profileCursor < len(m.profiles)-1 {
					m.profileCursor++
				}
			}
		case "tab", "left", "right":
			m.focusOnCLI = !m.focusOnCLI
		case "enter", " ":
			if len(m.profiles) == 0 {
				return m, nil
			}
			m.selectedProfile = m.profiles[m.profileCursor]
			m.selectedCLI = m.clis[m.cliCursor]
			return m, func() tea.Msg {
				return LaunchStartMsg{
					Profile: m.selectedProfile,
					CLI:     m.selectedCLI,
				}
			}
		case "esc", "q":
			return m, func() tea.Msg { return LaunchBackMsg{} }
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m LaunchModel) View() string {
	// Layout: 2 padding on each side
	sidePadding := 2
	contentWidth := m.width - sidePadding*2
	if contentWidth < 40 {
		contentWidth = 40
	}

	// Each pane takes half width
	paneWidth := contentWidth / 2

	// Calculate pane height (reserve 1 for help bar at bottom)
	paneHeight := m.height - 1
	if paneHeight < 10 {
		paneHeight = 10
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	// Build left pane content (Profiles)
	var leftContent strings.Builder
	leftContent.WriteString(titleStyle.Render("Select Profile"))
	leftContent.WriteString("\n")
	leftContent.WriteString(labelStyle.Render("Choose a profile to use"))
	leftContent.WriteString("\n\n")

	for i, p := range m.profiles {
		line := p
		pc := config.GetProfileConfig(p)
		if pc != nil && len(pc.Providers) > 0 {
			line += dimStyle.Render(fmt.Sprintf(" (%d)", len(pc.Providers)))
		}

		if i == m.profileCursor {
			if !m.focusOnCLI {
				leftContent.WriteString(selectedStyle.Render("> " + line))
			} else {
				leftContent.WriteString(itemStyle.Render("* " + line))
			}
		} else {
			leftContent.WriteString(itemStyle.Render("  " + line))
		}
		leftContent.WriteString("\n")
	}

	// Build right pane content (CLI)
	var rightContent strings.Builder
	rightContent.WriteString(titleStyle.Render("Select CLI"))
	rightContent.WriteString("\n")
	rightContent.WriteString(labelStyle.Render("Choose which CLI to launch"))
	rightContent.WriteString("\n\n")

	cliDescriptions := map[string]string{
		"claude":   "Claude Code",
		"codex":    "Codex CLI",
		"opencode": "OpenCode",
	}
	for i, c := range m.clis {
		line := c
		if desc, ok := cliDescriptions[c]; ok {
			line = desc + dimStyle.Render(" ("+c+")")
		}

		if i == m.cliCursor {
			if m.focusOnCLI {
				rightContent.WriteString(selectedStyle.Render("> " + line))
			} else {
				rightContent.WriteString(itemStyle.Render("* " + line))
			}
		} else {
			rightContent.WriteString(itemStyle.Render("  " + line))
		}
		rightContent.WriteString("\n")
	}

	// Pane style with thick border
	// Width is the internal content width (excluding border and padding)
	// Border takes 2 chars, padding takes 4 chars (2 each side)
	internalWidth := paneWidth - 2 - 4
	if internalWidth < 20 {
		internalWidth = 20
	}

	paneStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(internalWidth).
		Height(paneHeight - 2). // -2 for border top/bottom
		Padding(1, 2)

	leftPane := paneStyle.Render(leftContent.String())
	rightPane := paneStyle.Render(rightContent.String())

	// Join panes side by side
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Build the view line by line
	var view strings.Builder

	// Add side padding to each line of main content
	lines := strings.Split(mainContent, "\n")
	for _, line := range lines {
		view.WriteString(strings.Repeat(" ", sidePadding))
		view.WriteString(line)
		view.WriteString("\n")
	}

	// Fill remaining space to push help bar to bottom
	currentLines := len(lines)
	remainingLines := m.height - currentLines - 1
	for i := 0; i < remainingLines; i++ {
		view.WriteString("\n")
	}

	// Help bar at bottom - full terminal width
	helpBar := RenderHelpBar("Tab/←→ switch pane • ↑↓ navigate • Enter launch • Esc back", m.width)
	view.WriteString(helpBar)

	return view.String()
}

// Refresh reloads profiles and CLIs.
func (m *LaunchModel) Refresh() {
	m.profiles = config.ListProfiles()
	defaultProfile := config.GetDefaultProfile()
	defaultCLI := config.GetDefaultCLI()

	// Reset cursors to defaults
	for i, p := range m.profiles {
		if p == defaultProfile {
			m.profileCursor = i
			break
		}
	}
	for i, c := range m.clis {
		if c == defaultCLI {
			m.cliCursor = i
			break
		}
	}

	m.selectedProfile = defaultProfile
	m.selectedCLI = defaultCLI
	m.focusOnCLI = false
}
