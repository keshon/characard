package menu

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var asciiBanner = "" +
	"   ____ _                          ____               _ \n" +
	"  / ___| |__   __ _ _ __ __ _     / ___|__ _ _ __ ___| |\n" +
	" | |   | '_ \\ / _` | '__/ _` |   | |   / _` | '__/ _ | |\n" +
	" | |___| | | | (_| | | | (_| |   | |__| (_| | | | (_)| |\n" +
	"  \\____|_| |_|\\__,_|_|  \\__,_|    \\____\\__,_|_|  \\___|_|\n"

// Action is what the user selected from the main menu.
type Action string

const (
	ActionQuit     Action = "quit"
	ActionCards    Action = "cards"
	ActionSettings Action = "settings"
	ActionHelp     Action = "help"
	ActionAbout    Action = "about"
)

type menuItem struct {
	action Action
	label  string
}

var menuItems = []menuItem{
	{ActionCards, "Cards    — Manage character cards (list, select, batch rename)"},
	{ActionSettings, "Settings — Path to card folder"},
	{ActionHelp, "Help     — Usage and commands"},
	{ActionAbout, "About   — Credits"},
	{ActionQuit, "Quit"},
}

// Model is the main menu TUI model.
type Model struct {
	selectedIndex int
	width         int
	height        int
}

// SelectedAction returns the action that was chosen (set when user presses Enter).
func (m *Model) SelectedAction() Action {
	if m.selectedIndex < 0 || m.selectedIndex >= len(menuItems) {
		return ActionQuit
	}
	return menuItems[m.selectedIndex].action
}

// NewModel returns an initial menu model.
func NewModel() Model {
	return Model{selectedIndex: 0}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.selectedIndex = len(menuItems) - 1
			return m, tea.Quit
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
			return m, nil
		case "down", "j":
			if m.selectedIndex < len(menuItems)-1 {
				m.selectedIndex++
			}
			return m, nil
		case "enter", " ":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	var b strings.Builder

	bannerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true)
	bannerBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(0, 2).
		Margin(0, 0, 1, 0)
	header := bannerBox.Render(bannerStyle.Render(asciiBanner))
	subHeader := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("Character Card TUI Manager")
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(subHeader)
	b.WriteString("\n\n")

	menuTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Choose an option:")
	b.WriteString(menuTitle)
	b.WriteString("\n\n")

	selStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	normStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	arrow := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("▸")

	for i, item := range menuItems {
		if i == m.selectedIndex {
			b.WriteString("  ")
			b.WriteString(arrow)
			b.WriteString(" ")
			b.WriteString(selStyle.Render(item.label))
		} else {
			b.WriteString("    ")
			b.WriteString(normStyle.Render(item.label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("↑/↓ move · Enter select · q quit"))
	return b.String()
}

// Run runs the main menu and returns the selected action.
func Run() (Action, error) {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return ActionQuit, err
	}
	model, ok := final.(Model)
	if !ok {
		return ActionQuit, nil
	}
	return model.SelectedAction(), nil
}

// HelpText returns the full help content for the Help screen.
func HelpText() string {
	return `Character Card TUI Manager

COMMANDS (you can also run these directly):
  charcard                    Start interactive menu (this screen)
  charcard tui [dir]          Interactive card manager (default: from Settings)
    -dir <path>               Directory containing PNG character cards
  charcard settings list      Show current settings (cards folder path)
  charcard settings set-cards-dir -dir <path>   Set cards folder
  charcard help               This help
  charcard about              Credits

WORKFLOW:
  1. Run "Settings" to set the cards folder (or "charcard settings set-cards-dir -dir ./cards").
  2. Run "Cards" (or "charcard tui") to open the card list.
  3. Use Space to select/deselect, + to select all, - to clear selection.
  4. Press Enter with at least one card selected to choose an action:
     - Batch Rename: enter a template (e.g. [name]-[tag1]-[year].png), live preview, confirm.
  5. Template fields: [name], [creator], [tag1], [tag2], [year], [month], [index].
`
}

// AboutText returns the about content.
func AboutText() string {
	return `Character Card TUI Manager

CLI/TUI for PNG character cards (spec v1/v2/v3).
Extract metadata, list cards, batch rename by template.
Built with Bubble Tea.
`
}
