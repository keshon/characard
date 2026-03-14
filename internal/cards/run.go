package cards

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/keshon/charapng"
)

// Run starts the cards TUI for the given directory. It scans for PNG character
// cards, then runs the interactive list (select, batch rename). Returns when
// the user quits or on error.
func Run(ctx context.Context, dir string) error {
	files, err := charapng.ScanDirectory(dir)
	if err != nil {
		return fmt.Errorf("scan %s: %w", dir, err)
	}
	m := NewModel(ctx, dir, files)
	pm := newProgramModel(m)
	p := tea.NewProgram(pm, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
