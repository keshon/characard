package menu

import (
	"path/filepath"
	"strings"

	"github.com/keshon/characard/internal/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type settingsModel struct {
	width    int
	height   int
	cardsDir string
	input    textinput.Model
	saved    bool
	err      string
}

func newSettingsModel() (settingsModel, error) {
	s, err := config.GetSettings()
	if err != nil {
		return settingsModel{}, err
	}
	dir := s.CardsDir
	if dir != "" {
		dir = filepath.Clean(dir)
	}
	ti := textinput.New()
	ti.Placeholder = "e.g. ./cards or C:\\CharacterCards"
	ti.Width = 50
	ti.SetValue(dir)
	ti.Focus()
	return settingsModel{
		cardsDir: dir,
		input:    ti,
	}, nil
}

func (m settingsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m settingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, tea.Quit
		case "enter":
			dir := strings.TrimSpace(m.input.Value())
			if dir == "" {
				m.err = "Path cannot be empty"
				return m, nil
			}
			dir = filepath.Clean(dir)
			s, _ := config.GetSettings()
			s.CardsDir = dir
			if err := config.SetSettings(s); err != nil {
				m.err = err.Error()
				return m, nil
			}
			m.saved = true
			m.err = ""
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m settingsModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Settings")
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Cards folder — directory containing PNG character cards (OS-agnostic path)")
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Cards folder:")
	back := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter save · Esc back")
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(sub)
	b.WriteString("\n\n")
	b.WriteString(label)
	b.WriteString("\n")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")
	if m.err != "" {
		b.WriteString(errStyle.Render(m.err))
		b.WriteString("\n\n")
	}
	b.WriteString(back)
	return b.String()
}

// RunSettings shows the settings TUI (cards folder path) and returns when the user saves or exits.
func RunSettings() error {
	model, err := newSettingsModel()
	if err != nil {
		return err
	}
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
