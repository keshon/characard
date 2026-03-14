package help

import (
	"context"

	"github.com/keshon/characard/internal/menu"

	"github.com/keshon/commandkit"
)

func init() {
	commandkit.DefaultRegistry.Register(&helpCommand{})
}

type helpCommand struct{}

func (c *helpCommand) Name() string        { return "help" }
func (c *helpCommand) Description() string { return "Show usage and commands" }

func (c *helpCommand) Run(ctx context.Context, inv *commandkit.Invocation) error {
	return menu.RunHelp()
}
