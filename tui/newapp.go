package tui

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dopejs/opencc/internal/config"
)

// AppScreen represents the current screen in the new TUI.
type AppScreen int

const (
	ScreenMenu AppScreen = iota
	ScreenDashboard
	ScreenSettings
)

// NewAppModel is the main application model for the new TUI.
type NewAppModel struct {
	screen    AppScreen
	menu      MenuModel
	dashboard DashboardModel
	settings  SettingsModel
	width     int
	height    int
}

// NewNewAppModel creates a new application model.
func NewNewAppModel() NewAppModel {
	return NewAppModel{
		screen:    ScreenMenu,
		menu:      NewMenuModel(),
		dashboard: NewDashboardModel(),
		settings:  NewSettingsModel(),
	}
}

// Init implements tea.Model.
func (m NewAppModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m NewAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.screen {
	case ScreenMenu:
		return m.updateMenu(msg)
	case ScreenDashboard:
		return m.updateDashboard(msg)
	case ScreenSettings:
		return m.updateSettings(msg)
	}

	return m, nil
}

func (m NewAppModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case MenuSelectedMsg:
		switch msg.Action {
		case MenuLaunch:
			// TODO: Launch wizard
			return m, tea.Quit
		case MenuConfigure:
			m.screen = ScreenDashboard
			m.dashboard.Refresh()
			return m, nil
		case MenuSettings:
			m.screen = ScreenSettings
			m.settings.Refresh()
			return m, m.settings.Init()
		case MenuWebUI:
			// Open web UI in browser
			port := config.GetWebPort()
			url := fmt.Sprintf("http://127.0.0.1:%d", port)
			openBrowser(url)
			return m, nil
		case MenuQuit:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m NewAppModel) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case DashboardBackMsg:
		m.screen = ScreenMenu
		m.menu.Refresh()
		return m, nil
	case DashboardEditProviderMsg, DashboardEditProfileMsg, DashboardAddProviderMsg, DashboardAddProfileMsg:
		// TODO: Integrate with existing editors or create new ones
		return m, nil
	}

	var cmd tea.Cmd
	m.dashboard, cmd = m.dashboard.Update(msg)
	return m, cmd
}

func (m NewAppModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case SettingsSavedMsg, SettingsCancelledMsg:
		m.screen = ScreenMenu
		m.menu.Refresh()
		return m, nil
	}

	var cmd tea.Cmd
	m.settings, cmd = m.settings.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m NewAppModel) View() string {
	switch m.screen {
	case ScreenMenu:
		return m.menu.View()
	case ScreenDashboard:
		return m.dashboard.View()
	case ScreenSettings:
		return m.settings.View()
	}
	return ""
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("cmd", "/c", "start", url).Start()
	}
}

// RunNewApp runs the new TUI application.
func RunNewApp() error {
	p := tea.NewProgram(NewNewAppModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
