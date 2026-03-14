package settings

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/keshon/characard/internal/config"

	"github.com/keshon/commandkit"
)

func init() {
	commandkit.DefaultRegistry.Register(&settingsCommand{})
	commandkit.DefaultRegistry.Register(&settingsListCommand{})
	commandkit.DefaultRegistry.Register(&settingsSetCardsDirCommand{})
}

// --- settings (parent) ---

type settingsCommand struct{}

func (c *settingsCommand) Name() string { return "settings" }
func (c *settingsCommand) Description() string {
	return "View and change app settings (cards folder path)"
}

func (c *settingsCommand) Run(ctx context.Context, inv *commandkit.Invocation) error {
	return fmt.Errorf("use: settings list | settings set-cards-dir -dir <path>")
}

// --- settings list ---

type settingsListCommand struct{}

func (c *settingsListCommand) Name() string        { return "settings-list" }
func (c *settingsListCommand) Description() string { return "Show current settings" }

func (c *settingsListCommand) Run(ctx context.Context, inv *commandkit.Invocation) error {
	s, err := config.GetSettings()
	if err != nil {
		return err
	}
	path, _ := config.SettingsPath()
	fmt.Println("Settings file:", path)
	fmt.Println()
	cardsDir := s.CardsDir
	if cardsDir == "" {
		cardsDir = "(not set — set via 'charcard settings set-cards-dir -dir <path>')"
	} else {
		cardsDir = filepath.Clean(cardsDir)
	}
	fmt.Println("Cards folder:", cardsDir)
	return nil
}

// --- settings set-cards-dir ---

type settingsSetCardsDirCommand struct {
	dir string
}

func (c *settingsSetCardsDirCommand) Name() string { return "settings-set-cards-dir" }
func (c *settingsSetCardsDirCommand) Description() string {
	return "Set the folder containing PNG character cards"
}

func (c *settingsSetCardsDirCommand) Flags(fs *flag.FlagSet) {
	fs.StringVar(&c.dir, "dir", "", "Path to the cards folder (OS-agnostic)")
}

func (c *settingsSetCardsDirCommand) Run(ctx context.Context, inv *commandkit.Invocation) error {
	dir := c.dir
	if dir == "" && len(inv.Args) > 0 {
		dir = inv.Args[0]
	}
	if dir == "" {
		return fmt.Errorf("cards directory required: charcard settings set-cards-dir -dir <path>")
	}
	dir = filepath.Clean(dir)
	s, err := config.GetSettings()
	if err != nil {
		return err
	}
	s.CardsDir = dir
	if err := config.SetSettings(s); err != nil {
		return err
	}
	fmt.Printf("Cards folder set to: %s\n", dir)
	return nil
}
