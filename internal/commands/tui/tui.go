package tui

import (
	"context"
	"flag"

	"github.com/keshon/commandkit"

	"github.com/keshon/characard/internal/cards"
	"github.com/keshon/characard/internal/cli"
	"github.com/keshon/characard/internal/config"
)

func init() {
	commandkit.DefaultRegistry.Register(&tuiCommand{})
}

type tuiCommand struct {
	dir string
}

func (c *tuiCommand) Name() string { return "tui" }
func (c *tuiCommand) Description() string {
	return "Interactive character card manager (list, select, batch rename)"
}

func (c *tuiCommand) Flags(fs *flag.FlagSet) {
	fs.StringVar(&c.dir, "dir", ".", "Directory containing PNG character cards")
}

func (c *tuiCommand) Run(ctx context.Context, inv *commandkit.Invocation) error {
	dir := c.dir
	if data := cli.CLIDataFrom(inv); data != nil && len(data.Args) > 0 {
		dir = data.Args[0]
	}
	if dir == "" || dir == "." {
		if s, err := config.GetSettings(); err == nil && s.CardsDir != "" {
			dir = s.CardsDir
		} else {
			dir = "."
		}
	}
	return cards.Run(ctx, dir)
}
