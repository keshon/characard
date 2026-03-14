package menu

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type aboutModel struct {
	width  int
	height int
}

func (m aboutModel) Init() tea.Cmd { return nil }

func (m aboutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m aboutModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("About")
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Press any key to go back")
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(AboutText()))
	b.WriteString("\n\n")
	b.WriteString(sub)
	return b.String()
}

// RunAbout shows the about content and returns when the user presses a key.
func RunAbout() error {
	p := tea.NewProgram(aboutModel{}, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
