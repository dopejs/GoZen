package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/gozen/internal/config"
)

type fallbackModel struct {
	profile    string   // profile name ("default", "work", etc.)
	allConfigs []string // all available providers
	order      []string // current fallback order (selected providers)
	cursor     int      // cursor position in allConfigs
	grabbed    bool     // true = item is grabbed and arrow keys reorder
	standalone bool     // true = standalone CLI mode (no routing section)

	// Routing section
	section         int                                 // 0=default providers, 1=routing scenarios
	routingCursor   int                                 // cursor in routing scenarios
	routingExpanded map[config.Scenario]bool            // which scenarios are expanded
	routingOrder    map[config.Scenario][]string        // provider order per scenario
	routingModels   map[config.Scenario]map[string]string // per-provider models per scenario

	status string
	saved  bool // true = save succeeded, waiting to exit
}

func newFallbackModel(profile string) fallbackModel {
	if profile == "" {
		profile = "default"
	}
	return fallbackModel{
		profile:         profile,
		routingExpanded: make(map[config.Scenario]bool),
		routingOrder:    make(map[config.Scenario][]string),
		routingModels:   make(map[config.Scenario]map[string]string),
	}
}

type fallbackLoadedMsg struct {
	allConfigs []string
	order      []string
	routing    map[config.Scenario]*config.ScenarioRoute
}

func (m fallbackModel) init() tea.Cmd {
	profile := m.profile
	return func() tea.Msg {
		names := config.ProviderNames()
		pc := config.GetProfileConfig(profile)
		var order []string
		var routing map[config.Scenario]*config.ScenarioRoute
		if pc != nil {
			order = pc.Providers
			routing = pc.Routing
		}
		return fallbackLoadedMsg{allConfigs: names, order: order, routing: routing}
	}
}

func (m fallbackModel) update(msg tea.Msg) (fallbackModel, tea.Cmd) {
	// After save, ignore everything except saveExitMsg
	if m.saved {
		if _, ok := msg.(saveExitMsg); ok {
			return m, func() tea.Msg { return switchToListMsg{} }
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case fallbackLoadedMsg:
		m.allConfigs = msg.allConfigs
		m.order = msg.order
		m.cursor = 0
		// Load routing data
		if msg.routing != nil {
			for scenario, route := range msg.routing {
				m.routingOrder[scenario] = route.ProviderNames()
				m.routingModels[scenario] = make(map[string]string)
				for _, pr := range route.Providers {
					if pr.Model != "" {
						m.routingModels[scenario][pr.Name] = pr.Model
					}
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// orderIndex returns the 1-based index in order, or 0 if not in order.
func (m fallbackModel) orderIndex(name string) int {
	for i, n := range m.order {
		if n == name {
			return i + 1
		}
	}
	return 0
}

func (m fallbackModel) handleKey(msg tea.KeyMsg) (fallbackModel, tea.Cmd) {
	if m.grabbed {
		return m.handleGrabbed(msg)
	}

	switch msg.String() {
	case "esc", "q":
		// Cancel — return without saving
		return m, func() tea.Msg { return switchToListMsg{} }
	case "tab":
		// Switch between sections (skip in standalone mode)
		if !m.standalone {
			if m.section == 0 {
				m.section = 1
				m.routingCursor = 0
			} else {
				m.section = 0
				m.cursor = 0
			}
		}
	case "up", "k":
		if m.section == 0 {
			if m.cursor > 0 {
				m.cursor--
			}
		} else {
			if m.routingCursor > 0 {
				m.routingCursor--
			}
		}
	case "down", "j":
		if m.section == 0 {
			maxPos := len(m.allConfigs) - 1
			if m.standalone {
				maxPos = len(m.allConfigs) // +1 for save button
			}
			if m.cursor < maxPos {
				m.cursor++
			}
		} else {
			if m.routingCursor < len(knownScenarios)-1 {
				m.routingCursor++
			}
		}
	case " ":
		if m.section == 0 {
			// Toggle selection in default providers
			if m.cursor < len(m.allConfigs) {
				name := m.allConfigs[m.cursor]
				if idx := m.orderIndex(name); idx > 0 {
					// Remove from order
					m.order = removeFromOrder(m.order, name)
				} else {
					// Add to end of order
					m.order = append(m.order, name)
				}
			}
		}
	case "enter":
		if m.section == 0 {
			// Save button (standalone mode)
			if m.standalone && m.cursor == len(m.allConfigs) {
				return m.saveAndExit()
			}
			// Enter grab mode only if current item is in order
			if m.cursor < len(m.allConfigs) {
				name := m.allConfigs[m.cursor]
				if m.orderIndex(name) > 0 {
					m.grabbed = true
				}
			}
		} else {
			// Toggle scenario expansion or enter scenario editor
			if m.routingCursor < len(knownScenarios) {
				scenario := knownScenarios[m.routingCursor].scenario
				// Enter scenario editor
				return m, func() tea.Msg {
					return switchToScenarioEditMsg{
						profile:  m.profile,
						scenario: scenario,
					}
				}
			}
		}
	case "ctrl+s", "cmd+s":
		return m.saveAndExit()
	}
	return m, nil
}

func (m fallbackModel) saveAndExit() (fallbackModel, tea.Cmd) {
	pc := &config.ProfileConfig{
		Providers: m.order,
	}

	// Build routing config
	if len(m.routingOrder) > 0 {
		pc.Routing = make(map[config.Scenario]*config.ScenarioRoute)
		for scenario, providerNames := range m.routingOrder {
			if len(providerNames) == 0 {
				continue
			}
			var providerRoutes []*config.ProviderRoute
			for _, name := range providerNames {
				pr := &config.ProviderRoute{Name: name}
				if models, ok := m.routingModels[scenario]; ok {
					if model, ok := models[name]; ok && model != "" {
						pr.Model = model
					}
				}
				providerRoutes = append(providerRoutes, pr)
			}
			pc.Routing[scenario] = &config.ScenarioRoute{Providers: providerRoutes}
		}
	}

	if err := config.SetProfileConfig(m.profile, pc); err != nil {
		m.status = "Error: " + err.Error()
		return m, nil
	}
	m.saved = true
	m.status = "Saved"
	return m, saveExitTick()
}

func (m fallbackModel) handleGrabbed(msg tea.KeyMsg) (fallbackModel, tea.Cmd) {
	if m.cursor >= len(m.allConfigs) {
		m.grabbed = false
		return m, nil
	}
	name := m.allConfigs[m.cursor]
	orderIdx := m.orderIndex(name)
	if orderIdx == 0 {
		m.grabbed = false
		return m, nil
	}

	switch msg.String() {
	case "esc", "enter":
		m.grabbed = false
	case "up", "k":
		// Move up in order (swap with previous in order)
		if orderIdx > 1 {
			m.order[orderIdx-1], m.order[orderIdx-2] = m.order[orderIdx-2], m.order[orderIdx-1]
		}
	case "down", "j":
		// Move down in order (swap with next in order)
		if orderIdx < len(m.order) {
			m.order[orderIdx-1], m.order[orderIdx] = m.order[orderIdx], m.order[orderIdx-1]
		}
	}
	return m, nil
}

func removeFromOrder(order []string, name string) []string {
	var result []string
	for _, n := range order {
		if n != name {
			result = append(result, n)
		}
	}
	return result
}

// RunEditProfile is the standalone entry point for editing a profile.
func RunEditProfile(profile string) error {
	fm := newFallbackModel(profile)
	fm.standalone = true
	wrapper := &standaloneFallbackModel{fallback: fm}
	p := tea.NewProgram(wrapper, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return err
	}
	sm := result.(*standaloneFallbackModel)
	if sm.cancelled {
		return fmt.Errorf("cancelled")
	}
	if sm.fallback.saved {
		fmt.Printf("Profile %q updated.\n", profile)
	}
	return nil
}

// RunAddProfile is the standalone entry point for creating a new profile.
// Name input and provider selection are on a single page.
func RunAddProfile(presetName string) error {
	fm := newFallbackModel("__new__")
	fm.standalone = true

	ti := textinput.New()
	ti.Placeholder = "profile name"
	ti.Prompt = ""
	ti.CharLimit = 64
	if presetName != "" {
		ti.SetValue(presetName)
	}
	ti.Focus()

	wrapper := &standaloneFallbackModel{
		fallback:    fm,
		isNew:       true,
		nameInput:   ti,
		nameFocused: true,
	}
	p := tea.NewProgram(wrapper, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return err
	}
	sm := result.(*standaloneFallbackModel)
	if sm.cancelled {
		return fmt.Errorf("cancelled")
	}
	if sm.fallback.saved {
		fmt.Printf("Profile %q created.\n", sm.fallback.profile)
	}
	return nil
}

// standaloneFallbackModel wraps fallbackModel for standalone CLI use.
// It uses viewClean (no border, no routing) and quits immediately on save.
// When isNew=true, it also shows a name input field for creating a new profile.
type standaloneFallbackModel struct {
	fallback    fallbackModel
	cancelled   bool
	isNew       bool            // true = creating new profile
	nameInput   textinput.Model // name input (only when isNew)
	nameFocused bool            // true = name input has focus
	nameErr     string          // validation error for name
}

func (w *standaloneFallbackModel) Init() tea.Cmd {
	cmds := []tea.Cmd{w.fallback.init()}
	if w.isNew {
		cmds = append(cmds, textinput.Blink)
	}
	return tea.Batch(cmds...)
}

func (w *standaloneFallbackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			w.cancelled = true
			return w, tea.Quit
		}
		// When creating: route keys based on focus
		if w.isNew && w.nameFocused {
			return w.updateNameInput(msg)
		}
		// In provider list: shift+tab goes back to name input
		// Also: up at first provider goes back to name input
		if w.isNew && (msg.String() == "shift+tab" ||
			((msg.String() == "up" || msg.String() == "k") && w.fallback.cursor == 0 && !w.fallback.grabbed)) {
			w.nameFocused = true
			w.nameInput.Focus()
			return w, textinput.Blink
		}
		// Intercept save button (enter on save position) to validate name first
		if w.isNew && msg.String() == "enter" && w.fallback.cursor == len(w.fallback.allConfigs) {
			name := strings.TrimSpace(w.nameInput.Value())
			if name == "" {
				w.nameErr = "name is required"
				w.nameFocused = true
				w.nameInput.Focus()
				return w, textinput.Blink
			}
			for _, p := range config.ListProfiles() {
				if p == name {
					w.nameErr = fmt.Sprintf("profile %q already exists", name)
					w.nameFocused = true
					w.nameInput.Focus()
					return w, textinput.Blink
				}
			}
			w.nameErr = ""
			w.fallback.profile = name
			// Fall through to normal save handling
		}
	case switchToListMsg:
		// esc in provider list = cancel
		if !w.isNew {
			// Edit mode: clean exit
			return w, tea.Quit
		}
		w.cancelled = true
		return w, tea.Quit
	}

	var cmd tea.Cmd
	w.fallback, cmd = w.fallback.update(msg)

	// Quit immediately on save
	if w.fallback.saved {
		return w, tea.Quit
	}

	return w, cmd
}

func (w *standaloneFallbackModel) updateNameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		w.cancelled = true
		return w, tea.Quit
	case "tab", "down":
		w.nameErr = ""
		w.nameFocused = false
		w.nameInput.Blur()
		return w, nil
	case "enter":
		name := strings.TrimSpace(w.nameInput.Value())
		if name == "" {
			w.nameErr = "name is required"
			return w, nil
		}
		w.nameErr = ""
		w.nameFocused = false
		w.nameInput.Blur()
		return w, nil
	}
	var cmd tea.Cmd
	w.nameInput, cmd = w.nameInput.Update(msg)
	return w, cmd
}

func (w *standaloneFallbackModel) View() string {
	if w.isNew {
		return w.viewCreate()
	}
	return w.fallback.viewClean()
}

func (w *standaloneFallbackModel) viewCreate() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Create New Profile"))
	b.WriteString("\n\n")

	// Name input
	if w.nameFocused {
		b.WriteString("  Name: ")
		b.WriteString(w.nameInput.View())
	} else {
		name := w.nameInput.Value()
		if name == "" {
			name = "(empty)"
		}
		b.WriteString(dimStyle.Render("  Name: " + name))
	}
	b.WriteString("\n")
	if w.nameErr != "" {
		b.WriteString(errorStyle.Render("  " + w.nameErr))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Hint
	if w.nameFocused {
		b.WriteString(dimStyle.Render("  Enter/Tab to select providers"))
	} else {
		b.WriteString(dimStyle.Render("  Space toggle, Enter reorder, Esc cancel"))
	}
	b.WriteString("\n\n")

	// Provider list
	fm := w.fallback
	if len(fm.allConfigs) == 0 {
		b.WriteString(dimStyle.Render("  No providers configured.\n"))
		b.WriteString(dimStyle.Render("  Run 'zen config add provider' to create one."))
		return b.String()
	}

	for i, name := range fm.allConfigs {
		cursor := "  "
		style := dimStyle
		if !w.nameFocused && i == fm.cursor {
			cursor = "▸ "
			style = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
		}

		orderIdx := fm.orderIndex(name)
		var checkbox string
		if orderIdx > 0 {
			checkbox = lipgloss.NewStyle().
				Foreground(successColor).
				Render(fmt.Sprintf("[%d]", orderIdx))
		} else {
			checkbox = dimStyle.Render("[ ]")
		}

		grabIndicator := ""
		if fm.grabbed && !w.nameFocused && i == fm.cursor {
			grabIndicator = " " + lipgloss.NewStyle().
				Foreground(accentColor).
				Render("(reordering)")
		}

		line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, name, grabIndicator)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Save button
	b.WriteString("\n")
	if !w.nameFocused && fm.cursor == len(fm.allConfigs) {
		b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("▸ [ Save ]"))
	} else {
		b.WriteString(dimStyle.Render("  [ Save ]"))
	}

	if fm.status != "" && !fm.saved {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render("  ✗ " + fm.status))
	}

	return b.String()
}

// viewClean renders a borderless provider-toggle list for standalone use.
// No routing section, no border, no help bar.
func (m fallbackModel) viewClean() string {
	var b strings.Builder

	title := fmt.Sprintf("Profile: %s", m.profile)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Space toggle, Enter reorder, Esc cancel"))
	b.WriteString("\n\n")

	if len(m.allConfigs) == 0 {
		b.WriteString(dimStyle.Render("  No providers configured.\n"))
		b.WriteString(dimStyle.Render("  Run 'zen config add provider' to create one."))
		return b.String()
	}

	for i, name := range m.allConfigs {
		cursor := "  "
		style := dimStyle
		if i == m.cursor {
			cursor = "▸ "
			style = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
		}

		orderIdx := m.orderIndex(name)
		var checkbox string
		if orderIdx > 0 {
			checkbox = lipgloss.NewStyle().
				Foreground(successColor).
				Render(fmt.Sprintf("[%d]", orderIdx))
		} else {
			checkbox = dimStyle.Render("[ ]")
		}

		grabIndicator := ""
		if m.grabbed && i == m.cursor {
			grabIndicator = " " + lipgloss.NewStyle().
				Foreground(accentColor).
				Render("(reordering)")
		}

		line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, name, grabIndicator)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Save button
	b.WriteString("\n")
	if m.cursor == len(m.allConfigs) {
		b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("▸ [ Save ]"))
	} else {
		b.WriteString(dimStyle.Render("  [ Save ]"))
	}

	if m.status != "" && !m.saved {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render("  ✗ " + m.status))
	}

	return b.String()
}

func (m fallbackModel) view(width, height int) string {
	// Use global layout dimensions
	contentWidth, _, _, _ := LayoutDimensions(width, height)
	sidePadding := 2

	var b strings.Builder

	// Header
	title := "Group: default"
	if m.profile != "" && m.profile != "default" {
		title = fmt.Sprintf("Group: %s", m.profile)
	}
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("📦 " + title)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Content box with proper width
	boxWidth := contentWidth * 60 / 100
	if boxWidth < 50 {
		boxWidth = 50
	}
	if boxWidth > 80 {
		boxWidth = 80
	}

	var content strings.Builder
	if len(m.allConfigs) == 0 {
		content.WriteString(dimStyle.Render("No providers configured.\n"))
		content.WriteString(dimStyle.Render("Run 'zen config add provider' to create one."))
	} else {
		// Default Providers Section
		sectionStyle := sectionTitleStyle
		if m.section != 0 {
			sectionStyle = dimStyle
		}
		content.WriteString(sectionStyle.Render(" Default Providers"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" Space to toggle, Enter to reorder"))
		content.WriteString("\n\n")

		for i, name := range m.allConfigs {
			cursor := "  "
			style := tableRowStyle
			if m.section == 0 && i == m.cursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}

			orderIdx := m.orderIndex(name)
			var checkbox string
			if orderIdx > 0 {
				checkbox = lipgloss.NewStyle().
					Foreground(successColor).
					Render(fmt.Sprintf("[%d]", orderIdx))
			} else {
				checkbox = dimStyle.Render("[ ]")
			}

			grabIndicator := ""
			if m.grabbed && m.section == 0 && i == m.cursor {
				grabIndicator = " " + lipgloss.NewStyle().
					Foreground(accentColor).
					Render("(reordering)")
			}

			line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, name, grabIndicator)
			content.WriteString(style.Render(line))
			if i < len(m.allConfigs)-1 {
				content.WriteString("\n")
			}
		}

		// Routing Section
		content.WriteString("\n\n")
		sectionStyle = sectionTitleStyle
		if m.section != 1 {
			sectionStyle = dimStyle
		}
		content.WriteString(sectionStyle.Render(" Scenario Routing"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" Enter to configure scenario"))
		content.WriteString("\n\n")

		for i, ks := range knownScenarios {
			cursor := "  "
			style := tableRowStyle
			if m.section == 1 && i == m.routingCursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}

			// Check if configured
			providerCount := 0
			if order, ok := m.routingOrder[ks.scenario]; ok && len(order) > 0 {
				providerCount = len(order)
			}

			// Show provider count if configured
			countInfo := ""
			if providerCount > 0 {
				countInfo = dimStyle.Render(fmt.Sprintf(" (%d providers)", providerCount))
			}

			line := fmt.Sprintf("%s%s%s", cursor, ks.label, countInfo)
			content.WriteString(style.Render(line))
			if i < len(knownScenarios)-1 {
				content.WriteString("\n")
			}
		}
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(boxWidth).
		Render(content.String())

	b.WriteString(contentBox)
	b.WriteString("\n\n")

	if m.saved {
		b.WriteString(successStyle.Render("  ✓ " + m.status))
	} else {
		if m.status != "" {
			b.WriteString(errorStyle.Render("  ✗ " + m.status))
			b.WriteString("\n")
		}
	}

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
	helpBar := RenderHelpBar("Tab switch section • s save • Esc back", width)
	view.WriteString(helpBar)

	return view.String()
}
