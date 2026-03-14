package menu

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpModel struct {
	width  int
	height int
}

func (m helpModel) Init() tea.Cmd { return nil }

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "enter", "esc", " ":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m helpModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Help")
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Press any key to go back")
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(HelpText()))
	b.WriteString("\n\n")
	b.WriteString(sub)
	return b.String()
}

// RunHelp shows the help content and returns when the user presses a key.
func RunHelp() error {
	p := tea.NewProgram(helpModel{}, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
